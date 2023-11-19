package fontscan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	meta "github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
)

type cacheEntry struct {
	Location

	Family string
	meta.Aspect
}

// Logger is a type that can log warnings.
type Logger interface {
	Printf(format string, args ...interface{})
}

// The family substitution algorithm is copied from fontconfig
// and the match algorithm is inspired from Rust font-kit library

// FontMap provides a mechanism to select a [font.Face] from a font description.
// It supports system and user-provided fonts, and implements the CSS font substitutions
// rules.
//
// Note that [FontMap] is NOT safe for concurrent use, but several font maps may coexist
// in an application.
//
// [FontMap] is designed to work with an index built by scanning the system fonts,
// which is a costly operation (see [UseSystemFonts] for more details).
// A lightweight alternative is provided by the [FindFont] function, which only uses
// file paths to select a font.
type FontMap struct {
	logger Logger
	// caches of already loaded faceCache : the two maps are updated conjointly
	firstFace font.Face
	faceCache map[Location]font.Face
	metaCache map[font.Font]cacheEntry

	// the database to query, either loaded from an index
	// or populated with the [UseSystemFonts], [AddFont], and/or [AddFace] method.
	database  fontSet
	scriptMap map[language.Script][]int
	lru       runeLRU

	// built holds whether the candidates are populated.
	built bool
	// the candidates for the current query, which influences ResolveFace output
	candidates candidates

	// internal buffers used in SetQuery
	footprintsBuffer scoredFootprints
	cribleBuffer     familyCrible

	query Query // current query
}

// NewFontMap return a new font map, which should be filled with the `UseSystemFonts`
// or `AddFont` methods. The provided logger will be used to record non-fatal errors
// encountered during font loading. If logger is nil, log.Default() is used.
func NewFontMap(logger Logger) *FontMap {
	if logger == nil {
		logger = log.New(log.Writer(), "fontscan", log.Flags())
	}
	fm := &FontMap{
		logger:       logger,
		faceCache:    make(map[Location]font.Face),
		metaCache:    make(map[font.Font]cacheEntry),
		cribleBuffer: make(familyCrible),
		scriptMap:    make(map[language.Script][]int),
	}
	fm.lru.maxSize = 4096
	return fm
}

// SetRuneCacheSize configures the size of the cache powering [FontMap.ResolveFace].
// Applications displaying large quantities of text should tune this value to be greater
// than the number of unique glyphs they expect to display at one time in order to achieve
// optimal performance when segmenting text by face rune coverage.
func (fm *FontMap) SetRuneCacheSize(size int) {
	fm.lru.maxSize = size
}

// UseSystemFonts loads the system fonts and adds them to the font map.
// This method is safe for concurrent use, but should only be called once
// per font map.
// The first call of this method trigger a rather long scan.
// A per-application on-disk cache is used to speed up subsequent initialisations.
// Callers can provide an appropriate directory path within which this cache may be
// stored. If the empty string is provided, the FontMap will attempt to infer a correct,
// platform-dependent cache path.
//
// NOTE: On Android, callers *must* provide a writable path manually, as it cannot
// be inferred without access to the Java runtime environment of the application.
func (fm *FontMap) UseSystemFonts(cacheDir string) error {
	// safe for concurrent use; subsequent calls are no-ops
	err := initSystemFonts(fm.logger, cacheDir)
	if err != nil {
		return err
	}

	// systemFonts is read-only, so may be used concurrently
	fm.appendFootprints(systemFonts.flatten()...)

	fm.built = false

	fm.lru.Clear()
	return nil
}

// appendFootprints adds the provided footprints to the database and maps their script
// coverage.
func (fm *FontMap) appendFootprints(footprints ...footprint) {
	startIdx := len(fm.database)
	fm.database = append(fm.database, footprints...)
	// Insert entries into scriptMap for each footprint's covered scripts.
	for i, fp := range footprints {
		dbIdx := startIdx + i
		for _, script := range fp.scripts {
			fm.scriptMap[script] = append(fm.scriptMap[script], dbIdx)
		}
	}
}

// systemFonts is a global index of the system fonts.
// initSystemFontsOnce protects the initial assignment,
// and `systemFonts` use is then read-only
var (
	systemFonts         systemFontsIndex
	initSystemFontsOnce sync.Once
)

func cacheDir(userProvided string) (string, error) {
	if userProvided != "" {
		return userProvided, nil
	}
	// load an existing index
	if runtime.GOOS == "android" {
		// There is no stable way to infer the proper place to store the cache
		// with access to the Java runtime for the application. Rather than
		// clutter our API with that, require the caller to provide a path.
		return "", fmt.Errorf("user must provide cache directory on android")
	}
	configDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolving index cache path: %s", err)
	}
	return configDir, nil
}

// initSystemFonts scan the system fonts and update `SystemFonts`.
// If the returned error is nil, `SystemFonts` is guaranteed to contain
// at least one valid font.Face.
// It is protected by sync.Once, and is then safe to use by multiple goroutines.
func initSystemFonts(logger Logger, userCacheDir string) error {
	var err error

	initSystemFontsOnce.Do(func() {
		const cacheFilePattern = "font_index_v%d.cache"

		// load an existing index
		var dir string
		dir, err = cacheDir(userCacheDir)
		if err != nil {
			return
		}

		cachePath := filepath.Join(dir, fmt.Sprintf(cacheFilePattern, cacheFormatVersion))

		systemFonts, err = refreshSystemFontsIndex(logger, cachePath)
	})

	return err
}

func refreshSystemFontsIndex(logger Logger, cachePath string) (systemFontsIndex, error) {
	fontDirectories, err := DefaultFontDirectories(logger)
	if err != nil {
		return nil, fmt.Errorf("searching font directories: %s", err)
	}
	logger.Printf("using system font dirs %q", fontDirectories)

	currentIndex, _ := deserializeIndexFile(cachePath)
	// if an error occured (the cache file does not exists or is invalid), we start from scratch

	updatedIndex, err := scanFontFootprints(logger, currentIndex, fontDirectories...)
	if err != nil {
		return nil, fmt.Errorf("scanning system fonts: %s", err)
	}

	// since ResolveFace must always return a valid face, we make sure
	// at least one font exists and is valid.
	// Otherwise, the font map is useless; this is an extreme case anyway.
	err = updatedIndex.assertValid()
	if err != nil {
		return nil, fmt.Errorf("loading system fonts: %s", err)
	}

	// write back the index in the cache file
	err = updatedIndex.serializeToFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("updating cache: %s", err)
	}

	return updatedIndex, nil
}

// [AddFont] loads the faces contained in [fontFile] and add them to
// the font map.
// [fileID] is used as the [Location.File] entry returned by [FontLocation].
//
// If `familyName` is not empty, it is used as the family name for `fontFile`
// instead of the one found in the font file.
//
// An error is returned if the font resource is not supported.
//
// The order of calls to [AddFont] and [AddFace] determines relative priority
// of manually loaded fonts. See [ResolveFace] for details about when this matters.
func (fm *FontMap) AddFont(fontFile font.Resource, fileID, familyName string) error {
	loaders, err := loader.NewLoaders(fontFile)
	if err != nil {
		return fmt.Errorf("unsupported font resource: %s", err)
	}

	// eagerly load the faces
	faces, err := font.ParseTTC(fontFile)
	if err != nil {
		return fmt.Errorf("unsupported font resource: %s", err)
	}

	// by construction of fonts.Loader and fonts.FontDescriptor,
	// fontDescriptors and face have the same length
	if len(faces) != len(loaders) {
		panic("internal error: inconsistent font descriptors and loader")
	}

	var addedFonts []footprint
	for i, fontDesc := range loaders {
		fp, _, err := newFootprintFromLoader(fontDesc, true, scanBuffer{})
		// the font won't be usable, just ignore it
		if err != nil {
			continue
		}

		fp.Location.File = fileID
		fp.Location.Index = uint16(i)
		// TODO: for now, we do not handle variable fonts

		if familyName != "" {
			// give priority to the user provided family
			fp.Family = meta.NormalizeFamily(familyName)
		}

		addedFonts = append(addedFonts, fp)
		fm.cache(fp, faces[i])
	}

	if len(addedFonts) == 0 {
		return fmt.Errorf("empty font resource %s", fileID)
	}

	fm.appendFootprints(addedFonts...)

	fm.built = false

	fm.lru.Clear()
	return nil
}

// [AddFace] inserts an already-loaded font.Face into the FontMap. The caller
// is responsible for ensuring that [md] is accurate for the face.
//
// The order of calls to [AddFont] and [AddFace] determines relative priority
// of manually loaded fonts. See [ResolveFace] for details about when this matters.
func (fm *FontMap) AddFace(face font.Face, md meta.Description) {
	fp := newFootprintFromFont(face.Font, md)
	fm.cache(fp, face)

	fm.appendFootprints(fp)

	fm.built = false
	fm.lru.Clear()
}

func (fm *FontMap) cache(fp footprint, face font.Face) {
	if fm.firstFace == nil {
		fm.firstFace = face
	}
	fm.faceCache[fp.Location] = face
	fm.metaCache[face.Font] = cacheEntry{fp.Location, fp.Family, fp.Aspect}
}

// FontLocation returns the origin of the provided font. If the font was not
// previously returned from this FontMap by a call to ResolveFace, the zero
// value will be returned instead.
func (fm *FontMap) FontLocation(ft font.Font) Location {
	return fm.metaCache[ft].Location
}

// FontMetadata returns a description of the provided font. If the font was not
// previously returned from this FontMap by a call to ResolveFace, the zero
// value will be returned instead.
func (fm *FontMap) FontMetadata(ft font.Font) (family string, aspect meta.Aspect) {
	item := fm.metaCache[ft]
	return item.Family, item.Aspect
}

// SetQuery set the families and aspect required, influencing subsequent
// `ResolveFace` calls.
func (fm *FontMap) SetQuery(query Query) {
	if len(query.Families) == 0 {
		query.Families = []string{""}
	}
	fm.query = query
	fm.built = false
}

// candidates is a cache storing the indices into FontMap.database of footprints matching a Query
type candidates struct {
	// the two fallback slices have the same length: the number of family in the query
	withFallback    [][]int // for each queried family
	withoutFallback []int   // for each queried family, only one footprint is selected

	manual []int // manually inserted faces to be tried if the other candidates fail.
}

func (cd *candidates) resetWithSize(candidateSize int) {
	if cap(cd.withFallback) < candidateSize { // reallocate
		cd.withFallback = make([][]int, candidateSize)
		cd.withoutFallback = make([]int, candidateSize)
	}

	// only reslice
	cd.withFallback = cd.withFallback[0:candidateSize]
	cd.withoutFallback = cd.withoutFallback[0:candidateSize]

	// reset to "zero" values
	for i := range cd.withoutFallback {
		cd.withFallback[i] = nil
		cd.withoutFallback[i] = -1
	}
	cd.manual = cd.manual[0:]
}

func (fm *FontMap) buildCandidates() {
	if fm.built {
		return
	}
	fm.candidates.resetWithSize(len(fm.query.Families))

	selectFootprints := func(systemFallback bool) {
		for familyIndex, family := range fm.query.Families {
			candidates := fm.database.selectByFamily(family, systemFallback, &fm.footprintsBuffer, fm.cribleBuffer)
			if len(candidates) == 0 {
				continue
			}

			// select the correct aspects
			candidates = fm.database.retainsBestMatches(candidates, fm.query.Aspect)

			if systemFallback {
				// candidates is owned by fm.footprintsBuffer: copy its content
				S := fm.candidates.withFallback[familyIndex]
				if L := len(candidates); cap(S) < L {
					S = make([]int, L)
				} else {
					S = S[:L]
				}
				copy(S, candidates)
				fm.candidates.withFallback[familyIndex] = S
			} else {
				// when no systemFallback is required, the CSS spec says
				// that only one font among the candidates must be tried
				fm.candidates.withoutFallback[familyIndex] = candidates[0]
			}
		}
	}

	selectFootprints(false)
	selectFootprints(true)

	fm.candidates.manual = fm.database.filterUserProvided(fm.candidates.manual)
	fm.candidates.manual = fm.database.retainsBestMatches(fm.candidates.manual, fm.query.Aspect)
	fm.built = true
}

// returns nil if not candidates supports the rune `r`
func (fm *FontMap) resolveForRune(candidates []int, r rune) font.Face {
	// we first look up for an exact family match, without substitutions
	for _, footprintIndex := range candidates {
		// check the coverage
		if fp := fm.database[footprintIndex]; fp.Runes.Contains(r) {
			// try to use the font
			face, err := fm.loadFont(fp)
			if err != nil { // very unlikely; try an other family
				fm.logger.Printf("failed loading face: %v", err)
				continue
			}

			return face
		}
	}

	return nil
}

// ResolveFace select a font based on the current query (see `SetQuery`),
// and supporting the given rune, applying CSS font selection rules.
// The function will return nil if the underlying font database is empty,
// or if the file system is broken; otherwise the returned [font.Face] is always valid.
//
// If no fonts match the current query for the current rune according to the
// builtin matching process, the fonts added manually by [AddFont] and [AddFace]
// will be searched in the order in which they were added for a font with coverage
// for the provided rune. The first font covering the requested rune will be returned.
//
// If no fonts match after the manual font search, an arbitrary face will be returned.
func (fm *FontMap) ResolveFace(r rune) (face font.Face) {
	key := fm.lru.KeyFor(fm.query, r)
	face, ok := fm.lru.Get(key, fm.query)
	if ok {
		return face
	}
	defer func() {
		fm.lru.Put(key, fm.query, face)
	}()
	// Build the candidates if we missed the cache. If they're already built this is a
	// no-op.
	fm.buildCandidates()
	// we first look up for an exact family match, without substitutions
	for _, footprintIndex := range fm.candidates.withoutFallback {
		if footprintIndex == -1 {
			continue
		}
		if face := fm.resolveForRune([]int{footprintIndex}, r); face != nil {
			return face
		}
	}

	// if no family has matched so far, try again with system fallback
	for _, footprintIndexList := range fm.candidates.withFallback {
		if face := fm.resolveForRune(footprintIndexList, r); face != nil {
			return face
		}
	}

	// try manually loaded faces even if the typeface doesn't match, looking for matching aspects
	// and rune coverage.
	for _, footprintIndex := range fm.candidates.manual {
		if footprintIndex == -1 {
			continue
		}
		if face := fm.resolveForRune([]int{footprintIndex}, r); face != nil {
			return face
		}
	}

	fm.logger.Printf("No font matched for %q and rune %U (%c) -> searching by script coverage and aspect", fm.query.Families, r, r)

	script := language.LookupScript(r)
	scriptCandidates, ok := fm.scriptMap[language.LookupScript(r)]
	if ok {
		aspectCandidates := make([]int, len(scriptCandidates))
		copy(aspectCandidates, scriptCandidates)
		// Filter candidates to those matching the requested aspect first.
		aspectCandidates = fm.database.retainsBestMatches(aspectCandidates, fm.query.Aspect)
		if face := fm.resolveForRune(aspectCandidates, r); face != nil {
			return face
		}
		fm.logger.Printf("No font matched for aspect %v, script %s, and rune %U (%c) -> searching by script coverage only", fm.query.Aspect, script, r, r)
		// aspectCandidates has been filtered down and has exactly enough excess capacity to hold
		// the other original candidates.
		allCandidates := aspectCandidates[len(aspectCandidates):len(aspectCandidates):cap(aspectCandidates)]
		// Populate allCandidates with every script candidate that isn't in aspectCandidates.
		for _, idx := range scriptCandidates {
			possibleIdx := sort.Search(len(aspectCandidates), func(i int) bool {
				return aspectCandidates[i] >= idx
			})
			if possibleIdx < len(aspectCandidates) && aspectCandidates[possibleIdx] == idx {
				continue
			}
			allCandidates = append(allCandidates, idx)
		}
		// Try allCandidates.
		if face := fm.resolveForRune(allCandidates, r); face != nil {
			return face
		}
	}

	fm.logger.Printf("No font matched for script %s and rune %U (%c) -> returning arbitrary face", script, r, r)
	// return an arbitrary face
	if fm.firstFace == nil && len(fm.database) > 0 {
		for _, fp := range fm.database {
			face, err := fm.loadFont(fp)
			if err != nil {
				// very unlikely; warn and keep going
				fm.logger.Printf("failed loading face: %v", err)
				continue
			}
			return face
		}
	}

	return fm.firstFace
	// refreshSystemFontsIndex makes sure at least one face is valid
	// and AddFont also check for valid font files, meaning that
	// a valid FontMap should always contain a valid face,
	// and we should never return a nil face.
}

func (fm *FontMap) loadFont(fp footprint) (font.Face, error) {
	if face, hasCached := fm.faceCache[fp.Location]; hasCached {
		return face, nil
	}

	// since user provided fonts are added to `fonts`
	// we may now assume the font is stored on the file system
	face, err := fp.loadFromDisk()
	if err != nil {
		return nil, err
	}

	// add the face to the cache
	fm.cache(fp, face)

	return face, nil
}
