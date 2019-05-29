package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spilliams/goraffe/pkg/guru"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var referrersCmd = &cobra.Command{
	Use:   "referrers <package.function>",
	Args:  validateReferrersArgs,
	Short: "Visualize function referrers",
	Long: `Visualize function referrers.

When you're in a top-level go repository (e.g. ~/go/src/github.com/spilliams/goraffe/),
and you run goraffe referrers foo.bar, this command searches your current directory
for a package named foo, and a function declaration within that package named bar.

The command will output a graph of all callers of that function, all the way up
to the root of the directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		logrus.Debugf("dir: %s", dir)

		seek := strings.Split(args[0], ".")
		position, err := findFuncDecl(dir, seek[0], seek[1])
		if err != nil {
			return err
		}

		logrus.Debugf("position: %#v", position)
		// offset is +5 because "func " is 5 long
		fpos := fmt.Sprintf("%s:#%d", position.Filename, position.Offset+5)
		return guru.Call("referrers", fpos)
	},
}

func findFuncDecl(dir, pkgName, funcName string) (*token.Position, error) {
	dir = path.Join(dir, pkgName)

	// parse the directory
	fset := token.FileSet{}
	pkgs, err := parser.ParseDir(&fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("packages: %#v", pkgs)

	//
	decl, err := exploreMap(pkgs, funcName)
	if err != nil {
		return nil, err
	}

	if decl == nil {
		return nil, fmt.Errorf("could not find a function declaraction `%s` in package %s", funcName, pkgName)
	}
	logrus.Debugf("decl: %v", decl)

	p := fset.Position(decl.Pos())
	return &p, nil
}

func exploreMap(pkgs map[string]*ast.Package, funcName string) (*ast.FuncDecl, error) {
	for name, pkg := range pkgs {
		logrus.Tracef("package %s", name)
		decl, err := explorePkg(pkg, funcName)
		if err != nil {
			logrus.Warnf("error exploring package %v", name)
			return nil, err
		}
		if decl != nil {
			return decl, nil
		}
	}
	return nil, nil
}

func explorePkg(pkg *ast.Package, funcName string) (*ast.FuncDecl, error) {
	for name, file := range pkg.Files {
		logrus.Tracef("  file %s", name)
		decl, err := exploreFile(file, funcName)
		if err != nil {
			logrus.Warnf("error exploring file %v", name)
			return nil, err
		}
		if decl != nil {
			return decl, nil
		}
	}
	return nil, nil
}

func exploreFile(file *ast.File, funcName string) (*ast.FuncDecl, error) {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			logrus.Tracef("    function %v (%v, %v)", funcDecl.Name, funcDecl.Pos(), funcDecl.End())
			if funcDecl.Name.String() == funcName {
				return funcDecl, nil
			}
		}
	}
	return nil, nil
}

func validateReferrersArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide at lease one argument, the package.func")
	}
	return nil
}
