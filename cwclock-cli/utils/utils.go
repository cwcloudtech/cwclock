package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const EMPTY = ""

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

func IsNotBlank(str string) bool {
	return len(str) > 0 && strings.TrimSpace(str) != EMPTY
}

func IsBlank(str string) bool {
	return !IsNotBlank(str)
}

// IsTrue return whether str is not a false value.
// False values are: false, no, off, ko, 0, and empty string.
func IsTrue(str string) bool {
	if IsBlank(str) {
		return false
	}

	normalized := strings.TrimSpace(strings.ToLower(str))

	falseValues := []string{"false", "ko", "no", "off", "0"}
	if slices.Contains(falseValues, normalized) {
		return false
	}

	if num, err := strconv.ParseFloat(normalized, 64); err == nil {
		return num > 0
	}

	return true
}

func IsEmpty(value interface{}) bool {
	if nil == value {
		return true
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return IsBlank(v.String())
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		zero := reflect.New(v.Type()).Elem()
		return reflect.DeepEqual(v.Interface(), zero.Interface())
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Invalid:
		return true
	default:
		return false
	}
}

func ForbidValue(value string, terms []string) bool {
	if IsBlank(value) {
		return true
	}

	for _, term := range terms {
		if value == term {
			return false
		}
	}

	return true
}

func ContainsValue(value string, terms []string) bool {
	if IsBlank(value) {
		return false
	}

	for _, term := range terms {
		if value == term {
			return true
		}
	}

	return false
}

func IsYes(value string) bool {
	return ContainsValue(value, []string{"y", "Y"})
}

func IsNotEmpty(value interface{}) bool {
	return !IsEmpty(value)
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func PromptUserForValue() string {
	reader := bufio.NewReader(os.Stdin)
	// ReadString will block until the delimiter is entered
	value, err := reader.ReadString('\n')
	if nil != err {
		fmt.Println("An error occured while reading input. Please try again", err)
		return EMPTY
	}

	// remove the delimeter from the string
	value = strings.TrimSuffix(value, "\n")
	return value
}

func PrintJson(class interface{}) {
	marchal_class, err := json.Marshal(class)
	ExitIfError(err)

	fmt.Printf("%s\n", JsonInlinePrint(string(marchal_class)))
}

func PrintHeader(class interface{}) {
	values := reflect.ValueOf(class)
	typesOf := values.Type()
	headerMsg := EMPTY
	for i := 0; i < values.NumField(); i++ {
		headerMsg = fmt.Sprintf("%s%s\t", headerMsg, typesOf.Field(i).Name)
	}

	fmt.Println(headerMsg)
}

func PrintPretty(firstLine string, class interface{}) {
	fmt.Printf("%s:\n", firstLine)

	v := reflect.ValueOf(class)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		if IsEmpty(fieldValue) {
			continue
		}

		fieldName := strings.Replace(t.Field(i).Name, "_", " ", -1)
		fmt.Printf("  ➤ %s: %s\n", fieldName, strings.TrimSpace(fmt.Sprintf("%v", fieldValue)))
	}
}

func PrintPrettyArray(firstLine string, lst []string) {
	fmt.Printf("%s:\n", firstLine)

	for _, elem := range lst {
		fmt.Printf("  ➤ %v\n", elem)
	}
}

func PrintArray(lst []string) {
	fmt.Printf("%s\n", strings.Join(lst, "\n"))
}

func PrintMultiRow(type_class interface{}, class interface{}) {
	PrintHeader(type_class)
	s := reflect.ValueOf(class)
	for i := 0; i < s.Len(); i++ {
		v := reflect.Indirect(s.Index(i))
		valuesMsg := EMPTY
		for i := 0; i < v.NumField(); i++ {
			valuesMsg = fmt.Sprintf("%s%v\t", valuesMsg, v.Field(i).Interface())
		}

		fmt.Println(valuesMsg)
	}
}

func PrintDynamicTable(rows []map[string]interface{}) {
	if len(rows) == 0 {
		fmt.Println("No hits found")
		return
	}

	columns := make([]string, 0)
	seen := make(map[string]struct{})
	for _, row := range rows {
		for key := range row {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			columns = append(columns, key)
		}
	}
	sort.Strings(columns)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, strings.Join(columns, "\t"))
	for _, row := range rows {
		values := make([]string, 0, len(columns))
		for _, column := range columns {
			values = append(values, formatDynamicValue(row[column]))
		}
		fmt.Fprintln(writer, strings.Join(values, "\t"))
	}
	writer.Flush()
}

func formatDynamicValue(value interface{}) string {
	if value == nil {
		return EMPTY
	}

	switch typedValue := value.(type) {
	case string:
		return typedValue
	case json.Number:
		return typedValue.String()
	case fmt.Stringer:
		return typedValue.String()
	}

	encodedValue, err := json.Marshal(value)
	if err == nil {
		return string(encodedValue)
	}

	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func JsonArrayPrint(lst []map[string]interface{}) string {
	encodedHits, err := json.Marshal(lst)
	ExitIfError(err)
	return string(encodedHits)
}

func JsonInlinePrint(in string) string {
	var out bytes.Buffer
	err := json.Compact(&out, []byte(in))
	if nil != err {
		return in
	}

	return out.String()
}

func JsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), EMPTY, "\t")
	if nil != err {
		return in
	}

	return out.String()
}

func ExitIfError(err error) {
	ExitIfErrorWithMsg("Error", err)
}

func ExitIfErrorWithouMsg(err error) {
	ExitIfNeeded(EMPTY, nil != err)
}

func ExitIfErrorWithMsg(msg string, err error) {
	ExitIfNeeded(fmt.Sprintf("%s: %s", msg, err), nil != err)
}

func ExitIfNeeded(msg string, exit bool) {
	if exit {
		if IsNotBlank(msg) {
			fmt.Println(msg)
		}
		os.Exit(1)
	}
}

func GetSystemEditor() string {
	editorCommand := os.Getenv("EDITOR")
	if IsBlank(editorCommand) {
		editorCommand = "vi"
	}

	return editorCommand
}

func ShortName(name string, hash string) string {
	if IsBlank(name) {
		return EMPTY
	}

	if IsBlank(hash) {
		lastDashIndex := strings.LastIndex(name, "-")
		if lastDashIndex != -1 {
			return name[:lastDashIndex]
		}
		return name
	}

	hashWithDash := "-" + hash
	if strings.HasSuffix(name, hashWithDash) {
		return name[:len(name)-len(hashWithDash)]
	}
	return name
}

func ParseJSONPayload(jsonStr string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}
	return result, nil
}
