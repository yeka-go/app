package openapi

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

func LoadYamlFile(file string) (MapSlice, error) {
	ms := MapSlice{}
	err := ms.FromFile(file)
	return ms, err
}

func LoadFromBytes(b []byte) (MapSlice, error) {
	ms := MapSlice{}
	err := ms.FromBytes(b)
	return ms, err
}

type MapSlice yaml.MapSlice

func (m MapSlice) PathExists(path string) bool {
	_, ok := m.GetPath(path)
	return ok
}

func (m MapSlice) GetPath(path string) (any, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var x any = yaml.MapSlice(m)

	if len(parts) == 1 && parts[0] == "" {
		return x, true
	}

	for _, p := range parts {
		p = strings.ReplaceAll(strings.ReplaceAll(p, "~1", "/"), "~0", "~")
		found := false
		switch xx := x.(type) {
		case yaml.MapSlice:
			for _, v := range xx {
				key, _ := v.Key.(string)
				if key != p {
					continue
				}
				x = v.Value
				found = true
				break
			}
		case []any:
			index, err := strconv.Atoi(p)
			if err != nil {
				return nil, false
			}

			for i := range xx {
				if i != index {
					continue
				}
				x = xx[i]
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return x, true
}

func (m MapSlice) GetPathAsString(path string) (string, bool) {
	v, ok := m.GetPath(path)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (m *MapSlice) AddPath(path string, obj any) error {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	lastPath := len(parts) == 1
	nextPath := ""
	if !lastPath {
		nextPath = "/" + strings.Join(parts[1:], "/")
	}

	if lastPath && parts[0] == "" {
		mobj, ok := obj.(yaml.MapSlice)
		if !ok {
			return errors.New("unable to add object")
		}
		return m.Append(mobj)
	}

	ms := *m
	for k, v := range ms {
		key, _ := v.Key.(string)
		if key != parts[0] {
			continue
		}
		key = strings.ReplaceAll(strings.ReplaceAll(key, "~1", "/"), "~0", "~")

		// if existing value is mapSlice, expect object is also map slice
		switch vv := v.Value.(type) {
		case yaml.MapSlice:
			if lastPath {
				mobj, ok := obj.(yaml.MapSlice)
				if !ok {
					return errors.New("invalid type to add to path")
				}

				vvm := MapSlice(vv)
				err := vvm.Append(mobj)
				if err != nil {
					return fmt.Errorf("error on \"%v\": %w", path, err)
				}
				vv = yaml.MapSlice(vvm) // TODO check for duplicate key between vv & mobj
			} else {
				nm := MapSlice(vv)
				err := nm.AddPath(nextPath, obj)
				if err != nil {
					return err
				}

				vv = yaml.MapSlice(nm)
			}

			v.Value = vv
			ms[k] = v
			*m = ms
			return nil

		case []any:
			mobj, ok := obj.([]any)
			if ok {
				vv = append(vv, mobj...)
			} else {
				vv = append(vv, obj)
			}
			v.Value = vv
			ms[k] = v
			*m = ms
			return nil
		}
		return errors.New("unable to add object")
	}

	if !lastPath {
		nm := MapSlice{}
		err := nm.AddPath(nextPath, obj)
		if err == nil {
			ms = append(ms, yaml.MapItem{Key: parts[0], Value: yaml.MapSlice(nm)})
			*m = ms
		}
		return err
	}

	ms = append(ms, yaml.MapItem{Key: parts[0], Value: obj})
	*m = ms
	return nil
}

func (m *MapSlice) SetPath(path string, newobj any) error {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var x any = (*yaml.MapSlice)(m)

	if len(parts) == 1 && parts[0] == "" {
		vv, ok := newobj.(yaml.MapSlice)
		if !ok {
			return errors.New("unable to replace value")
		}
		*m = MapSlice(vv)
		return nil
	}

	// slog.Debug("Parts", "parts", parts)
	for pi, p := range parts {
		p = strings.ReplaceAll(strings.ReplaceAll(p, "~1", "/"), "~0", "~")
		lastPart := len(parts)-1 == pi
		found := false
		// slog.Debug("  part: " + p)

		inf, isPointerInterface := x.(*any)
		if isPointerInterface {
			x = *inf
		}
		// msp, isMapSlicePointer := x.(*MapSlice)
		yms, isYamlMapSlice := x.(yaml.MapSlice)
		ymsp, isYamlMapSlicePointer := x.(*yaml.MapSlice)
		ma, isSlice := x.([]any)

		// slog.Debug("    parent data type:", "*any", isPointerInterface, "*yaml.MapSlice", isYamlMapSlicePointer, "yaml.MapSlice", isYamlMapSlice, "[]any", isSlice)

		/*
			if isMapSlicePointer {
				ms := *msp
				for i, v := range ms {
					key, _ := v.Key.(string)
					if key != p {
						continue
					}

					found = true
					if lastPart {
						v.Value = obj
						ms[i] = v
						*msp = ms
					} else {
						x = &ms[i].Value
					}
					break
				}
			} else
			//*/
		if isYamlMapSlice {
			for i, v := range yms {
				key, _ := v.Key.(string)
				if key != p {
					continue
				}

				found = true
				if lastPart {
					v.Value = newobj
					yms[i] = v
				} else {
					x = &yms[i].Value
				}
				break
			}
		} else if isYamlMapSlicePointer {
			ms := *ymsp
			for i, v := range ms {
				key, _ := v.Key.(string)
				if key != p {
					continue
				}

				found = true
				if lastPart {
					v.Value = newobj
					ms[i] = v
					*ymsp = ms
				} else {
					x = &ms[i].Value
				}
				break
			}
		} else if isSlice {
			index, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			for i := range ma {
				if i != index {
					continue
				}

				found = true
				if lastPart {
					ma[i] = newobj
				} else {
					x = ma[i]
				}
				break
			}
		} else {
			fmt.Printf("unknown type %#v\n", x)
		}

		if !found {
			// slog.Debug("    part not found", "part", p)
			var val any = &yaml.MapSlice{}
			if lastPart {
				val = newobj
			}
			newItem := &yaml.MapItem{Key: p, Value: val}
			// slog.Debug("    created", "newItem", newItem)

			if isYamlMapSlicePointer {
				ms := *ymsp
				ms = append(ms, *newItem)
				if !lastPart {
					x = &ms[len(ms)-1].Value
				}
				*ymsp = ms
			} else if isYamlMapSlice {
				yms = append(yms, *newItem)
				if isPointerInterface {
					*inf = yms
				}
				if !lastPart {
					x = &newItem.Value
				}
			} else {
				fmt.Printf("not found -> %#v\n", reflect.TypeOf(x).PkgPath())
			}
		}
	}
	return nil
}

/*
	func (m MapSlice) AddToPath(path string, obj any) error {
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		var x any = yaml.MapSlice(m)

		if len(parts) == 1 && parts[0] == "" {
			mobj, ok := obj.(yaml.MapSlice)
			if !ok {
				return errors.New("unable to add object")
			}
			return m.Append(mobj)
		}

		for _, p := range parts {
			found := false
			switch xx := x.(type) {
			case yaml.MapSlice:
				for _, v := range xx {
					key, _ := v.Key.(string)
					if key != p {
						continue
					}
					x = v.Value
					found = true
					break
				}
			case []any:
				index, err := strconv.Atoi(p)
				if err != nil {
					return nil, false
				}

				for i := range xx {
					if i != index {
						continue
					}
					x = xx[i]
					found = true
					break
				}
			}
			if !found {
				return nil, false
			}
		}
		return x, true
	}
*/

func (m *MapSlice) Append(obj yaml.MapSlice) error {
	keys := make(map[string]bool)
	for _, v := range *m {
		key, ok := v.Key.(string)
		if !ok {
			return fmt.Errorf("non string key on source: %#v", v.Key)
		}
		keys[key] = true
	}

	for _, v := range obj {
		key, ok := v.Key.(string)
		if !ok {
			return fmt.Errorf("non string key on new object: %#v", v.Key)
		}
		if keys[key] {
			return fmt.Errorf("key \"%v\" already exists", key)
		}
		*m = append(*m, v)
	}
	return nil
}

func (m MapSlice) ToYaml() ([]byte, error) {
	return yaml.MarshalWithOptions(
		yaml.MapSlice(m),
		yaml.IndentSequence(true),
		yaml.UseLiteralStyleIfMultiline(true),
	)
}

func (m *MapSlice) FromBytes(b []byte) error {
	var ms yaml.MapSlice
	err := yaml.UnmarshalWithOptions(b, &ms, yaml.UseOrderedMap())
	*m = MapSlice(ms)
	return err
}

func (m *MapSlice) FromFile(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return m.FromBytes(b)
}

// Ref contains information about $ref
type Ref struct {
	Path    string // Original path where $ref is reside (eg: "/component/schema/name")
	RefFile string // Referenced file ($ref: "file#/path")
	RefPath string // Referenced path inside the file ($ref: "file#/path")
}

func (m *MapSlice) FindRefs() []Ref {
	return findRefs([]string{}, yaml.MapSlice(*m))
}

func findRefs(path []string, m yaml.MapSlice) []Ref {
	res := make([]Ref, 0)
	for _, v := range m {
		key, ok := v.Key.(string)
		if !ok {
			log.Printf("key isn't string: %#v\n", v.Key)
			continue
		}

		newpath := append(path, strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1"))
		switch val := v.Value.(type) {
		case yaml.MapSlice:
			res = append(res, findRefs(newpath, val)...)
		case string:
			if key == "$ref" {
				part := strings.Split(val, "#")
				if len(part) == 1 {
					part = append(part, "/")
				} else if len(part) != 2 {
					log.Printf("invalid references on %v: %#v\n", strings.Join(newpath, "/"), val)
				}
				res = append(res, Ref{Path: "/" + strings.Join(newpath, "/"), RefFile: part[0], RefPath: part[1]})
			}
		case []any:
			res = append(res, findRefOnSlice(newpath, val)...)
		case bool, uint64, float64:
			// do nothing
		default:
			log.Printf("unknown value type on %v: %#V\n", strings.Join(newpath, "/"), v.Value)
		}
	}
	return res
}

func findRefOnSlice(path []string, m []any) []Ref {
	res := make([]Ref, 0)
	for i, v := range m {
		newpath := append(path, strconv.Itoa(i))
		switch val := v.(type) {
		case yaml.MapSlice:
			res = append(res, findRefs(newpath, val)...)
		case []any:
			res = append(res, findRefOnSlice(newpath, val)...)
		case string:
			// do nothing
		default:
			log.Printf("unknown value type on %v: %#v\n", strings.Join(newpath, "/"), v)
		}
	}
	return res
}
