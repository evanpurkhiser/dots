package installer

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/resolver"
)

type dotfileCompiler struct {
	dotfile  *resolver.Dotfile
	content  *bytes.Buffer
	compiled bool
	config   config.SourceConfig
	files    []*os.File
}

// ensureCompiled transforms the source dotfiles into the dotfile output. It
// will not recompile if the dotfile has already been compiled.
func (c *dotfileCompiler) ensureCompiled() error {
	if c.compiled {
		return nil
	}

	compiledData := []byte{}

	for i, sourceFile := range c.files {
		data, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return err
		}

		// 1. Always trim whitespace off of the source file.
		data = trimWhitespace(data)

		// 2. For any source file that procedes after the first, trim shebang
		//    markers for cleanlyness of bash configurations. We trim whitespace
		//    again to remove any space after the shebang.
		if i != 0 {
			data = trimShebang(data)
			data = trimWhitespace(data)
		}

		// Combine files with *one* blank line between them
		if i != 0 {
			compiledData = append(compiledData, '\n', '\n')
		}

		compiledData = append(compiledData, data...)
	}

	// 3. Expand environment variables if the dotfile was marked.
	if c.dotfile.ExpandEnv {
		compiledData = expandEnvironment(compiledData)
	}

	// 4. All files should end with a single newline
	compiledData = append(compiledData, '\n')

	// Store the compiled dotfile
	c.compiled = true
	c.content.Reset()
	c.content.Write(compiledData)

	return nil
}

// Read implements the io.Reader interface. Calling read will compile the
// dotfile into it's final byte slice.
func (c *dotfileCompiler) Read(p []byte) (int, error) {
	c.ensureCompiled()
	return c.content.Read(p)
}

// Close implments the io.Closer interface. Calling close will close all source
// files associated to the dotfile.
func (c *dotfileCompiler) Close() error {
	var err error

	for _, file := range c.files {
		closeErr := file.Close()
		if closeErr != nil {
			err = closeErr
		}
	}

	return err
}

// OpenDotfile opens a source dotfile for streaming compilation.
func OpenDotfile(dotfile *resolver.Dotfile, config config.SourceConfig) (io.ReadCloser, error) {
	files := make([]*os.File, len(dotfile.Sources))

	for i, source := range dotfile.Sources {
		file, err := os.Open(config.SourcePath + separator + source.Path)
		if err != nil {
			return nil, err
		}

		files[i] = file
	}

	compiler := &dotfileCompiler{
		dotfile: dotfile,
		content: bytes.NewBuffer(nil),
		config:  config,
		files:   files,
	}

	return compiler, nil
}

// mustCompile indicates if a dotfile must be compiled, or if a single-source
// dotfile does not require any transformations and may be directly installed.
func shouldCompile(dotfile *resolver.Dotfile, config config.SourceConfig) bool {
	if len(dotfile.Sources) > 1 {
		return true
	}

	if dotfile.ExpandEnv {
		return true
	}

	return false
}
