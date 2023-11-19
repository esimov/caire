package fontscan

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	meta "github.com/go-text/typesetting/opentype/api/metadata"
)

// defines the routines to serialize a font set to
// the disk

// assume len(dst) >= 4
func serializeFloat(f float32, dst []byte) {
	binary.BigEndian.PutUint32(dst, math.Float32bits(f))
}

// assume len(src) >= 4
func deserializeFloat(src []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(src))
}

func serializeString(s string) []byte {
	L := len(s)
	if L > math.MaxUint16 { // never happen in practice
		L = math.MaxUint16
	}
	buffer := make([]byte, 2+L) // len as uint16 + data
	binary.BigEndian.PutUint16(buffer, uint16(L))
	copy(buffer[2:], s)
	return buffer
}

func deserializeString(s *string, data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid string (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+L {
		return 0, errors.New("invalid string length (EOF)")
	}
	*s = string(data[2 : 2+L])
	return 2 + L, nil
}

const aspectSize = 1 + 4 + 4

// serializeTo serialize the Aspect in binary format
func serializeAspect(as meta.Aspect) []byte {
	var buffer [aspectSize]byte
	buffer[0] = byte(as.Style)
	serializeFloat(float32(as.Weight), buffer[1:])
	serializeFloat(float32(as.Stretch), buffer[5:])
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func deserializeAspectFrom(data []byte, as *meta.Aspect) (int, error) {
	if len(data) < aspectSize {
		return 0, errors.New("invalid Aspect (EOF)")
	}
	as.Style = meta.Style(data[0])
	as.Weight = meta.Weight(deserializeFloat(data[1:]))
	as.Stretch = meta.Stretch(deserializeFloat(data[5:]))
	return aspectSize, nil
}

// serializeTo serialize the Footprint in binary format,
// by appending to `dst` and returning the slice
func (fp footprint) serializeTo(dst []byte) []byte {
	dst = append(dst, serializeString(fp.Location.File)...)

	var buffer [4]byte
	binary.BigEndian.PutUint16(buffer[:], fp.Location.Index)
	binary.BigEndian.PutUint16(buffer[2:], fp.Location.Instance)
	dst = append(dst, buffer[:]...)

	dst = append(dst, serializeString(fp.Family)...)
	dst = append(dst, fp.Runes.serialize()...)
	dst = append(dst, fp.scripts.serialize()...)
	dst = append(dst, serializeAspect(fp.Aspect)...)

	return dst
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (fp *footprint) deserializeFrom(data []byte) (int, error) {
	n, err := deserializeString(&fp.Location.File, data)
	if err != nil {
		return 0, err
	}
	if len(data) < n+4 {
		return 0, errors.New("invalid Location (EOF)")
	}
	fp.Location.Index = binary.BigEndian.Uint16(data[n:])
	fp.Location.Instance = binary.BigEndian.Uint16(data[n+2:])
	n += 4

	read, err := deserializeString(&fp.Family, data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = fp.Runes.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = fp.scripts.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = deserializeAspectFrom(data[n:], &fp.Aspect)
	if err != nil {
		return 0, err
	}
	n += read

	return n, nil
}

// serialize into binary format, appending to `dst` and returning
// the updated slice
func serializeFootprintsTo(footprints []footprint, dst []byte) []byte {
	for _, fp := range footprints {
		dst = fp.serializeTo(dst)
	}
	return dst
}

// parses the format written by `serializeFootprints`
func deserializeFootprints(src []byte) (out []footprint, err error) {
	for totalRead := 0; totalRead < len(src); {
		var fp footprint
		read, err := fp.deserializeFrom(src[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("invalid footprints: %s", err)
		}
		totalRead += read

		out = append(out, fp)
	}

	return out, nil
}

func (ff fileFootprints) serializeTo(dst []byte) []byte {
	dst = append(dst, serializeString(ff.path)...)
	dst = append(dst, ff.modTime.serialize()...)
	// end by the variable length footprint list
	dst = serializeFootprintsTo(ff.footprints, dst)
	return dst
}

func (ff *fileFootprints) deserializeFrom(src []byte) error {
	n, err := deserializeString(&ff.path, src)
	if err != nil {
		return err
	}
	if len(src) < n+8 {
		return errors.New("invalid fileFootprints (EOF)")
	}
	ff.modTime.deserialize(src[n:])
	n += 8
	ff.footprints, err = deserializeFootprints(src[n:])
	if err != nil {
		return err
	}
	return nil
}

const cacheFormatVersion = 2

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

// serialize into binary format, compressed with gzip
func (index systemFontsIndex) serializeTo(w io.Writer) error {
	// version as uint16 + len as uint32 + somewhat the minimum size for a footprint
	buffer := make([]byte, 6, max(6, 4+len(index)*(aspectSize+1+2)))
	binary.BigEndian.PutUint16(buffer[:], cacheFormatVersion)
	binary.BigEndian.PutUint32(buffer[2:], uint32(len(index)))

	for _, ff := range index {
		// add buffer to store the length of the encoded fileFootprints,
		// needed when decoding from a stream
		n := len(buffer)
		buffer = append(buffer, make([]byte, 4)...)

		buffer = ff.serializeTo(buffer)

		size := len(buffer) - n - 4
		binary.BigEndian.PutUint32(buffer[n:], uint32(size))
	}
	wr := gzip.NewWriter(w)
	_, err := wr.Write(buffer)
	if err != nil {
		return fmt.Errorf("serializing font footprints: %s", err)
	}
	err = wr.Close()
	if err != nil {
		return fmt.Errorf("compressing serialized font footprints: %s", err)
	}
	return nil
}

// parses the format written by `fontIndex.serializeTo`
func deserializeIndex(src io.Reader) (systemFontsIndex, error) {
	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("invalid compressed index file: %s", err)
	}
	defer r.Close()

	var (
		buf    [6]byte
		out    systemFontsIndex
		buffer bytes.Buffer
	)

	// read the expected length
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("invalid index format: %s", err)
	}
	version := binary.BigEndian.Uint16(buf[:])
	if version != cacheFormatVersion {
		return nil, fmt.Errorf("different index version format: found %d", version)
	}
	L := binary.BigEndian.Uint32(buf[2:])
	for i := uint32(0); i < L; i++ {
		// size of the encoded footprint
		if _, err := io.ReadFull(r, buf[:4]); err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}
		size := binary.BigEndian.Uint32(buf[:4])
		// buffer the fileFootprints segment
		buffer.Reset()
		_, err := io.CopyN(&buffer, r, int64(size))
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}

		var fp fileFootprints
		err = fp.deserializeFrom(buffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}

		out = append(out, fp)
	}

	return out, nil
}

func deserializeIndexFile(cachePath string) (systemFontsIndex, error) {
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out, err := deserializeIndex(f)
	return out, err
}

func (index systemFontsIndex) serializeToFile(cachePath string) error {
	dir, _ := filepath.Split(cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("unable to create system font cache dir %q: %w", dir, err)
	}
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}

	err = index.serializeTo(f)
	if err != nil {
		return err
	}

	err = f.Close()
	return err
}
