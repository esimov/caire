// SPDX-License-Identifier: Unlicense OR MIT

package gl

import (
	"errors"
	"fmt"
	"strings"
)

func CreateProgram(ctx *Functions, vsSrc, fsSrc string, attribs []string) (Program, error) {
	vs, err := CreateShader(ctx, VERTEX_SHADER, vsSrc)
	if err != nil {
		return Program{}, err
	}
	defer ctx.DeleteShader(vs)
	fs, err := CreateShader(ctx, FRAGMENT_SHADER, fsSrc)
	if err != nil {
		return Program{}, err
	}
	defer ctx.DeleteShader(fs)
	prog := ctx.CreateProgram()
	if !prog.Valid() {
		return Program{}, errors.New("glCreateProgram failed")
	}
	ctx.AttachShader(prog, vs)
	ctx.AttachShader(prog, fs)
	for i, a := range attribs {
		ctx.BindAttribLocation(prog, Attrib(i), a)
	}
	ctx.LinkProgram(prog)
	if ctx.GetProgrami(prog, LINK_STATUS) == 0 {
		log := ctx.GetProgramInfoLog(prog)
		ctx.DeleteProgram(prog)
		return Program{}, fmt.Errorf("program link failed: %s", strings.TrimSpace(log))
	}
	return prog, nil
}

func CreateComputeProgram(ctx *Functions, src string) (Program, error) {
	cs, err := CreateShader(ctx, COMPUTE_SHADER, src)
	if err != nil {
		return Program{}, err
	}
	defer ctx.DeleteShader(cs)
	prog := ctx.CreateProgram()
	if !prog.Valid() {
		return Program{}, errors.New("glCreateProgram failed")
	}
	ctx.AttachShader(prog, cs)
	ctx.LinkProgram(prog)
	if ctx.GetProgrami(prog, LINK_STATUS) == 0 {
		log := ctx.GetProgramInfoLog(prog)
		ctx.DeleteProgram(prog)
		return Program{}, fmt.Errorf("program link failed: %s", strings.TrimSpace(log))
	}
	return prog, nil
}

func CreateShader(ctx *Functions, typ Enum, src string) (Shader, error) {
	sh := ctx.CreateShader(typ)
	if !sh.Valid() {
		return Shader{}, errors.New("glCreateShader failed")
	}
	ctx.ShaderSource(sh, src)
	ctx.CompileShader(sh)
	if ctx.GetShaderi(sh, COMPILE_STATUS) == 0 {
		log := ctx.GetShaderInfoLog(sh)
		ctx.DeleteShader(sh)
		return Shader{}, fmt.Errorf("shader compilation failed: %s", strings.TrimSpace(log))
	}
	return sh, nil
}

func ParseGLVersion(glVer string) (version [2]int, gles bool, err error) {
	var ver [2]int
	if _, err := fmt.Sscanf(glVer, "OpenGL ES %d.%d", &ver[0], &ver[1]); err == nil {
		return ver, true, nil
	} else if _, err := fmt.Sscanf(glVer, "WebGL %d.%d", &ver[0], &ver[1]); err == nil {
		// WebGL major version v corresponds to OpenGL ES version v + 1
		ver[0]++
		return ver, true, nil
	} else if _, err := fmt.Sscanf(glVer, "%d.%d", &ver[0], &ver[1]); err == nil {
		return ver, false, nil
	}
	return ver, false, fmt.Errorf("failed to parse OpenGL ES version (%s)", glVer)
}
