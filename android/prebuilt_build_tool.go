// Copyright 2020 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package android

import (
	"path"
	"path/filepath"
)

func init() {
	RegisterModuleType("prebuilt_build_tool", prebuiltBuildToolFactory)
}

type prebuiltBuildToolProperties struct {
	// Source file to be executed for this build tool
	Src *string `android:"path,arch_variant"`

	// Extra files that should trigger rules using this tool to rebuild
	Deps []string `android:"path,arch_variant"`
}

type prebuiltBuildTool struct {
	ModuleBase
	prebuilt Prebuilt

	properties prebuiltBuildToolProperties

	toolPath OptionalPath
}

func (t *prebuiltBuildTool) Name() string {
	return t.prebuilt.Name(t.ModuleBase.Name())
}

func (t *prebuiltBuildTool) Prebuilt() *Prebuilt {
	return &t.prebuilt
}

func (t *prebuiltBuildTool) DepsMutator(ctx BottomUpMutatorContext) {
	if t.properties.Src == nil {
		ctx.PropertyErrorf("src", "missing prebuilt source file")
	}
}

func (t *prebuiltBuildTool) GenerateAndroidBuildActions(ctx ModuleContext) {
	sourcePath := t.prebuilt.SingleSourcePath(ctx)
	installedPath := PathForModuleOut(ctx, t.ModuleBase.Name())
	deps := PathsForModuleSrc(ctx, t.properties.Deps)

	var relPath string
	if filepath.IsAbs(installedPath.String()) {
		relPath = filepath.Join(absSrcDir, sourcePath.String())
	} else {
		var err error
		relPath, err = filepath.Rel(path.Dir(installedPath.String()), sourcePath.String())
		if err != nil {
			ctx.ModuleErrorf("Unable to generate symlink between %q and %q: %s", installedPath.String(), sourcePath.String(), err)
		}
	}

	ctx.Build(pctx, BuildParams{
		Rule:      Symlink,
		Output:    installedPath,
		Input:     sourcePath,
		Implicits: deps,
		Args: map[string]string{
			"fromPath": relPath,
		},
	})

	t.toolPath = OptionalPathForPath(installedPath)
}

func (t *prebuiltBuildTool) HostToolPath() OptionalPath {
	return t.toolPath
}

var _ HostToolProvider = &prebuiltBuildTool{}

// prebuilt_build_tool is to declare prebuilts to be used during the build, particularly for use
// in genrules with the "tools" property.
func prebuiltBuildToolFactory() Module {
	module := &prebuiltBuildTool{}
	module.AddProperties(&module.properties)
	InitSingleSourcePrebuiltModule(module, &module.properties, "Src")
	InitAndroidArchModule(module, HostSupportedNoCross, MultilibFirst)
	return module
}
