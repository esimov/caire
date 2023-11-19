# Description and purpose of the package

This package provides a way to locate and load a `font.Font`, which is the
fundamental object needed by `go-text` for shaping and text rendering.

## Use case

This package may be used by UI toolkits and markup language renderers. Both use-cases may need to display large quantities of text of varying languages and writing systems, and want to make use of all available fonts, both packaged within the application and installed on the system. In both cases, content/UI authors provide hints about the fonts that they want chosen (family names, weights, styles, etc...) and want the closest available match to the requested properties.

## Overview of the API

The entry point of the library is the `FontMap` type. It should be created for each text shaping task and be filled either with system fonts (by calling `UseSystemFonts`) or with user-provided font files (using `AddFont`, `AddFace`), or both.
To leverage all the system fonts, the first usage of `UseSystemFonts` triggers a scan which builds a font index. Its content is saved on disk so that subsequent usage by the same app are not slowed down by this step.

Once initialized, the font map is used to select fonts matching a `Query` with `SetQuery`. A query is defined by one or several families and an `Aspect`, containining style, weight, stretchiness. Finally, the font map satisfies the `shaping.Fontmap` interface, so that is may be used with `shaping.SplitByFace`.

## Zoom on the implementation

### Font directories

Fonts are searched by walking the file system, in the folders returned by `DefaultFontDirectories`, which are platform dependent.
The current list is copied from [fontconfig](https://gitlab.freedesktop.org/fontconfig/fontconfig) and [go-findfont](github.com/flopp/go-findfont).

### Font family substitutions

A key concept of the implementation (inspired by [fontconfig](https://gitlab.freedesktop.org/fontconfig/fontconfig)) is the idea to enlarge the requested family with similar known families.
This ensure that suitable font fallbacks may be provided even if the required font is not available.
It is implemented by a list of susbtitutions, each of them having a test and a list of additions.

Simplified example : if the list of susbtitutions is

- Test: the input family is Arial, Addition: Arimo
- Test: the input family is Arimo, Addition: sans-serif
- Test: the input family is sans-serif, Addition: DejaVu Sans et Verdana

then,

- for the Arimo input family, [Arimo, sans-serif, DejaVu Sans, Verdana] would be matched
- for the Arial input family, [Arial, Arimo, sans-serif, DejaVu Sans, Verdana] would be matched

To respect the user request, the order of the list is significant (first entries have higher priority).

`FontMap.SetQuery` apply a list of hard-coded subsitutions, extracted from
Fontconfig configurations files.

### Style matching

`FontMap.SetQuery` takes an optional argument describing the style of
the required font (style, weight, stretchiness).

When no exact match is found, the [CSS font selection rules](https://drafts.csswg.org/css-fonts/#font-prop) are applied to return the closest match.
As an example, if the user asks for `(Italic, ExtraBold)` but only `(Normal, Bold)` and `(Oblique, Bold)`
are available, the `(Oblique, Bold)` would be returned.

### System font index

The `FontMap` type requires more information than the font paths to be able to quickly and accurately
match a font against family, aspect, and rune coverage query. This information is provided by a list of font summaries,
which are lightweight enough to be loaded and queried efficiently.

The initial scan required to build this index has a significant latency (say between 0.2 and 0.5 sec on a laptop).
Once the first scan has been done, however, the subsequent launches are fast : at the first call of `UseSystemFonts`, the index is loaded from an on-disk cache, and its integrity is checked against the
current file system state to detect font installation or suppression.
