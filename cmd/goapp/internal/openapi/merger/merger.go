package merger

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/yeka-go/app/cmd/goapp/internal/openapi"
)

type ref struct {
	openapi.Ref
	RefFullFile string
}

type docref struct {
	filename string
	spec     openapi.MapSlice
	refList  []ref // list of $ref (origin path, target file & target path) inside the file

	// list of paths as seen from different files or itself
	uniquePath       map[string]bool
	sortedUniquePath []string
}

func (d docref) Key() string {
	return d.filename
}

type Doc struct {
	openapi.Map[docref, string]
}

func Open(file string) ([]byte, error) {
	ms, err := openapi.LoadYamlFile(file)
	if err != nil {
		return nil, err
	}

	doc := Doc{}

	if err := doc.LoadRefs(file, ms); err != nil {
		return nil, err
	}
	for file := range doc.Items() {
		slices.SortFunc(file.refList, func(a, b ref) int {
			return strings.Compare(a.Path, b.Path)
		})

	}

	// for ref := range doc.Items() {
	// 	fmt.Println("File: ", ref.filename)
	// 	for _, v := range ref.refList {
	// 		fmt.Println(v.Path, v.RefFile, v.RefPath)
	// 	}
	// }

	err = doc.Resolve()
	if err != nil {
		log.Println("Resolve error:", err)
		return nil, err
	}
	res, err := doc.Index(0).spec.ToYaml()
	return res, err
}

func (doc *Doc) LoadRefs(file string, spec openapi.MapSlice) error {
	refs := spec.FindRefs()
	dr := docref{
		filename:         file,
		spec:             spec,
		refList:          make([]ref, len(refs)),
		uniquePath:       make(map[string]bool),
		sortedUniquePath: []string{},
	}
	doc.Push(dr, dr.filename)

	for i, v := range refs {
		reffile := resolvePath(file, v.RefFile)
		doc.Get(file).refList[i] = ref{Ref: v, RefFullFile: reffile}
		if v.RefFile == "" {
			doc.Get(file).uniquePath[v.RefPath] = true
			doc.Get(file).refList[i].RefFullFile = file
			continue
		}

		if !doc.Exists(reffile) {
			ms, err := openapi.LoadYamlFile(reffile)
			if err != nil {
				return fmt.Errorf("loading %v: %w", reffile, err)
			}
			if err := doc.LoadRefs(reffile, ms); err != nil {
				return err
			}
		}
		doc.Get(reffile).uniquePath[v.RefPath] = true
	}
	return nil
}

func (doc *Doc) Resolve() error {
	return doc.resolve("/", doc.Index(0).filename, "/", &Resolver{resolvedPaths: []string{}, alias: make(map[string]string)})
}

var unnamedCounter = 0

func (doc *Doc) resolve(basePath, targetFile, targetBasePath string, reslv *Resolver) error {
	slog.Debug("START - Looking for $ref", "basePath", basePath, "file", targetFile, "targetBasePath", targetBasePath)
	defer slog.Debug("ENDED")

	for _, v := range doc.Get(targetFile).refList {
		if !strings.HasPrefix(strings.TrimSuffix(v.Path, "/")+"/", strings.TrimSuffix(targetBasePath, "/")+"/") {
			continue
		}

		slog.Debug("  Found", "basePath", basePath, "path", v.Path, "$ref-file", v.RefFile, "$ref-path", v.RefPath)
		newpath := CombinePath(basePath, v.Path, targetBasePath)
		if v.RefFullFile == doc.Index(0).filename {
			log.Printf("Setting path %v\n", newpath)
			doc.Index(0).spec.SetPath(newpath+"/$ref", "#"+v.RefPath)
			continue
		}

		obj, ok := doc.Get(v.RefFullFile).spec.GetPath(v.RefPath)
		if !ok {
			slog.Error("  Unable to resolve", "$ref-info", v)
			return errors.New("unable to resolve " + v.RefFile + "#" + v.RefPath)
		}

		targetPath := newpath

		lastPartOfPath := Path(v.RefPath).LastPath()
		if lastPartOfPath == "" {
			unnamedCounter++
			lastPartOfPath = "UNNAMED" + strconv.Itoa(unnamedCounter)
		}

		// get naming should be separated from resolved-check
		var err error
		alias, found := reslv.isResolvedWithAlias(v.RefFullFile, v.RefPath, lastPartOfPath)

		componentPath := getComponentPaths(newpath)
		if (strings.HasPrefix(newpath, "/paths/") && componentPath != "") || (componentPath != "" && strings.HasPrefix(newpath, componentPath+"/")) {
			targetPath = componentPath + "/" + alias

			if componentPath == "/components/examples" && Path(newpath).LastPath() == "examples" {
				slog.Debug("    Moving example", "target", newpath, "value-from", v.RefFile+"#"+v.RefPath)
				err = doc.Index(0).spec.SetPath(newpath, obj)
				if err != nil {
					return err
				}
				slog.Debug("    Resolving example")
				if err := doc.resolve(newpath, v.RefFullFile, v.RefPath, reslv); err != nil {
					return err
				}
				continue
			}

			slog.Debug("    SetRef2", "target", newpath+"/$ref", "value", "#"+targetPath, "cp", componentPath, "np", newpath)
			err = doc.Index(0).spec.SetPath(newpath+"/$ref", "#"+targetPath)
			if err != nil {
				return err
			}

			if !found {
				slog.Debug("    Moving2", "target", targetPath, "value-from", v.RefFile+"#"+v.RefPath)
				err = doc.Index(0).spec.SetPath(targetPath, obj)
				if err != nil {
					return err
				}

				slog.Debug("    Resolving2 current $ref")
				err = doc.resolve(targetPath, v.RefFullFile, v.RefPath, reslv)
				if err != nil {
					return err
				}
			}
			// slog.Debug("    ", "resolved-path", reslv.resolvedPaths)
			continue
		}

		// if found {
		// 	err = doc.Index(0).spec.SetPath(newpath+"/$ref", "#"+targetPath)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		if !found {
			slog.Debug("    Moving", "target", targetPath, "value-from", v.RefFile+"#"+v.RefPath)
			if err := doc.Index(0).spec.SetPath(targetPath, obj); err != nil {
				return err
			}
			slog.Debug("    Resolving")
			if err := doc.resolve(targetPath, v.RefFullFile, v.RefPath, reslv); err != nil {
				return err
			}
		}
	}
	return nil
}

func CombinePath(base, addition, sub string) string {
	if !strings.HasPrefix(addition, sub) {
		return ""
	}
	addition = strings.TrimPrefix(addition, sub)
	if !strings.HasPrefix(addition, "/") {
		addition = "/" + addition
	}
	if base == "/" {
		base = ""
	}
	return base + strings.TrimSuffix(addition, "/$ref")
}

type Resolver struct {
	resolvedPaths []string
	alias         map[string]string
}

func (r *Resolver) isResolvedWithAlias(file, path, alias string) (string, bool) {
	for _, v := range r.resolvedPaths {
		if file+"#"+path+"/" == v {
			return r.alias[v], true
		}
		if strings.HasPrefix(file+"#"+path+"/", v) {
			return alias, true
		}
	}

	r.resolvedPaths = append(r.resolvedPaths, file+"#"+path+"/")
	if alias != "" {
		r.alias[file+"#"+path+"/"] = alias
	}
	return alias, false
}

// probably need to sort by resolved path first for each original items

// isResolved(components.yaml /components/responses/UnauthorizedError)
// -- find filename, check if anything have prefix of given parts
//    -- if contains /components/responses/BadError -> resolved = true
//    -- if contains /components/responses/ -> resolved = true
//    -- if contains /components/ -> resolved true
// if already resolved, proceed to map

// also need to keep track of resolved path to determine how we resolve the object (direct or under /components)
// /path/~1hello/$ref -> paths.yaml#/ = /path/~1hello
// inside paths.yaml
//   /get/$ref -> ops.yml#/ = /path/~1hello/get

type prefixedLog string

func (p prefixedLog) Println(str string, args ...any) {
	log.Printf(string(p)+str+"\n", args...)
}

var componentLists = []string{"schema", "schemas", "responses", "parameters", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks"}

func getComponentPaths(path string) string {
	parts := strings.Split(path, "/")
	slices.Reverse(parts)
	for _, v := range parts {
		if slices.Contains(componentLists, v) {
			if v == "schema" {
				return "/components/schemas"
			} else {
				return "/components/" + v
			}
		}
	}
	return ""
}

func resolvePath(file1, file2 string) string {
	// fmt.Println(file1, file2)

	dir1 := strings.Split(path.Dir(file1), "/")
	dir2 := strings.Split(path.Dir(file2), "/")
	for range len(dir2) {
		if dir2[0] == "." {
			dir2 = dir2[1:]
		}
	}

	i := 0
	for i = 0; i < len(dir2); i++ {
		if dir2[i] != ".." {
			break
		}
		if len(dir1) <= 0 {
			log.Fatal("error path resolver")
		}
		dir1 = dir1[0 : len(dir1)-1]
	}
	return "./" + strings.Join(append(dir1, dir2[i:]...), "/") + "/" + path.Base(file2)
}

type Path string

func (p Path) LastPath() string {
	parts := strings.Split(string(p), "/")
	return parts[len(parts)-1]
}
