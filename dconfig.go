// Package dconfig assists in reading configuration files with the
// common OPTION=VALUE format.
//
// Updated 2017-04-22
//
// IMPORTANT NOTE: The 2017-04-22 update changed the package's method of
// operation and the format of the API function calls, breaking all programs
// written before then. I'm sorry for your loss, but it works better now.
//
package dconfig

import("bufio"; "errors"; "fmt"; "os"; "regexp"; "strconv"; "strings")

// STRIP is passed as a flag to AddString() to indicate that whitespace
// should be trimmed from the ends of the value when read from a
// configuration file.
const STRIP     uint8 = 1

// UPPER is passed as a flag to AddString() to indicate that the value
// read from the configuration file should be converted to upper case.
const UPPER     uint8 = 2

// LOWER is passed as a flag to AddString() to indicate that the value
// read from the configuration file should be converted to lower case.
const LOWER     uint8 = 4

// UNSIGNED is passed as a flag to AddInt() and AddFloat() to indicate
// that any leading negative signs should be ignored when reading from
// a configuration file.
const UNSIGNED  uint8 = 8

// NONE is passed as a flag to AddXxx() to indicate none of the other
// flags apply. It is ALSO returned by OptionType() to indicate the
// given option isn't present. (It's also used internally.)
const NONE      uint8 = 0

// STRING, INT, FLOAT, and BOOL are returned by OptionType() to
// indicate which type of value is expected for a given option. They
// are also used internally.
const (
    BOOL    uint8 = 16
    STRING  uint8 = 128
    INT     uint8 = 64
    FLOAT   uint8 = 32
)

const all_types uint8 = STRING | INT | FLOAT | BOOL
const disallowed_str_opts   uint8 = UNSIGNED
const disallowed_int_opts   uint8 = STRIP | UPPER | LOWER
const disallowed_float_opts uint8 = STRIP | UPPER | LOWER

var str_map      map[string]*string
var int_map      map[string]*int
var float_map    map[string]*float64
var bool_map     map[string]*bool
var option_flags map[string]uint8

var comment_re   *regexp.Regexp
var nonblank_re  *regexp.Regexp
var option_re    *regexp.Regexp
var int_token    *regexp.Regexp
var uint_token   *regexp.Regexp
var float_token  *regexp.Regexp
var ufloat_token *regexp.Regexp

var boolean_trues  [6]string = [6]string{"1", "t", "true", "y", "yes", "+"}
var boolean_falses [7]string = [7]string{"0", "f", "false", "n", "no", "-", "nil"}

// Return true if value val has the bit for attribute attrib set.
//
func hasAttr(val, attrib uint8) bool {
    if val & attrib == 0 {
        return false
    } else {
        return true
    }
}

// Return the appropriate boolean type represented in the given token.
// Returns false and an error when token is unrecognized.
//
func boolRead(token string) (bool, error) {
    txt := strings.ToLower(token)
    for _, t := range boolean_trues {
        if txt == t {
            return true, nil
        }
    }
    for _, f := range boolean_falses {
        if txt == f {
            return false, nil
        }
    }
    err_str := fmt.Sprintf("\"%s\" is not a recognized boolean value.", token)
    return false, errors.New(err_str)
}

// Return the number of bits set in a given bitmask. Used to determine
// whether more than one of a set of mutually-exclusive options is set.
//
func sumOfBits(bmask uint8) uint8 {
    var bit uint8 = 1
    var sum uint8 = 0
    for i := 0; i < 8; i++ {
        if bmask & bit != 0 {
            sum += 1
        }
        bit = bit<<1
    }
    return sum
}

// Reset() clears all the configured options.
// If a package your program uses ALSO uses package dconfig, and they have one
// or more identical keys, this could cause weird behavior. To avoid this,
// your program should
// * call dconfig.Reset()
// * configure all options with the dconfig.AddXxx() functions
// * call dconfig.Configure()
// without doing anything else in between.
//
func Reset() {
    str_map = make(map[string]*string)
    int_map = make(map[string]*int)
    float_map = make(map[string]*float64)
    bool_map = make(map[string]*bool)
    option_flags = make(map[string]uint8)
}

// OptionType() returns the type of value associated with a given option
// name. Returned are one of the aforementioned constants (STRING, INT,
// FLOAT, BOOL); returns the constant NONE if opt isn't configured.
//
func OptionType(opt string) uint8 {
    var exists bool
    uname := strings.ToUpper(opt)
    _, exists = str_map[uname]
    if exists {
        return STRING
    }
    _, exists = int_map[uname]
    if exists {
        return INT
    }
    _, exists = float_map[uname]
    if exists {
        return FLOAT
    }
    _, exists = bool_map[uname]
    if exists {
        return BOOL
    }
    return NONE
}

func optionExists(opt string) bool {
    return OptionType(opt) != NONE
}

// Adds an option that will be parsed as an integer when read from the
// configuration file.
//
// It can take the NONE flag, or the UNSIGNED flag (in which case any
// leading minus signs will be ignored when converting into an int).
//
func AddInt(target *int, name string, flags uint8) error {
    if flags & disallowed_int_opts != 0 {
        return errors.New("unsupported flag for integer option type")
    }
    
    uname := strings.ToUpper(name)
    if optionExists(uname) {
        return errors.New(fmt.Sprintf("\"%s\" option already exists", uname))
    }
    
    int_map[uname] = target
    option_flags[uname] = flags | INT
    
    return nil
}

// Adds an option that will be parsed as a string when read from the
// configuration file.
//
// In addition to the NONE flag, it can take a combination of the following:
//  * STRIP -- leading and trailing whitespace will be trimmed
//  * UPPER -- will be converted to upper case
//  * LOWER -- will be converted to lower case
//
// Don't use the last two together.
//
func AddString(target *string, name string, flags uint8) error {
    if flags & disallowed_str_opts != 0 {
        return errors.New("unsupported flag for string option type")
    }

    uname := strings.ToUpper(name)
    if optionExists(uname) {
        return errors.New(fmt.Sprintf("\"%s\" option already exists", uname))
    }
    
    str_map[uname] = target
    option_flags[uname] = flags | STRING
    
    return nil
}

// Adds an option that will be parsed as a floating point number when
// read from the configuration file.
//
// It can take the NONE flag, or the UNSIGNED flag (in which case any
// leading minus sign will be ignored when converting into a float).
//
func AddFloat(target *float64, name string, flags uint8) error {
    if flags & disallowed_float_opts != 0 {
        return errors.New("unsupported flag for float option type")
    }

    uname := strings.ToUpper(name)
    if optionExists(uname) {
        return errors.New(fmt.Sprintf("\"%s\" option already exists", uname))
    }
    
    float_map[uname] = target
    option_flags[uname] = flags | FLOAT
    
    return nil
}

// Adds an option that will be parsed as a boolean when read from the
// configuration file. Accepts many varieties of true/false representations.
//
func AddBool(target *bool, name string) error {
    uname := strings.ToUpper(name)
    if optionExists(uname) {
        return errors.New(fmt.Sprintf("\"%s\" option already exists", uname))
    }
    
    bool_map[uname] = target
    option_flags[uname] = BOOL
    
    return nil
}

// setOption() is called by Configure() fore each line that matches the
// OPTION=value pattern. It updates the appropriate xxx_map[] for each
// extant OPTION with a well-formed value.
//
func setOption(name, value string, verbose bool) error {
    uname := strings.ToUpper(name)
    flags, exists := option_flags[uname]
    if !exists {
        err_str := fmt.Sprintf("unrecognized option \"%s\"", uname)
        if verbose {
            fmt.Fprintf(os.Stderr, "%s\n", err_str)
        }
        return errors.New(err_str)
    }
    
    if hasAttr(flags, STRING) {
        if hasAttr(flags, STRIP) {
            value = strings.TrimSpace(value)
        }
        if hasAttr(flags, LOWER) {
            value = strings.ToLower(value)
        } else if hasAttr(flags, UPPER) {
            value = strings.ToUpper(value)
        }
        *(str_map[uname]) = value
        return nil
        
    } else if hasAttr(flags, INT) {
        if hasAttr(flags, UNSIGNED) {
            value = uint_token.FindString(value)
        } else {
            value = int_token.FindString(value)
        }
        iv, err := strconv.Atoi(value)
        if err != nil {
            err_str := fmt.Sprintf("\"%s\" not a recognizable integer", value)
            if verbose {
                fmt.Fprintf(os.Stderr, "%s\n", err_str)
            }
            return errors.New(err_str)
        }
        *(int_map[uname]) = iv
        return nil
        
    } else if hasAttr(flags, FLOAT) {
        if hasAttr(flags, UNSIGNED) {
            value = ufloat_token.FindString(value)
        } else {
            value = float_token.FindString(value)
        }
        fv, err := strconv.ParseFloat(value, 64)
        if err != nil {
            err_str := fmt.Sprintf("\"%s\" not a recognizable float", value)
            if verbose {
                fmt.Fprintf(os.Stderr, "%s\n", err_str)
            }
            return errors.New(err_str)
        }
        *(float_map[uname]) = fv
        return nil
        
    } else if hasAttr(flags, BOOL) {
        value = strings.TrimSpace(value)
        value = strings.ToLower(value)
        for _, t := range boolean_trues {
            if value == t {
                *(bool_map[uname]) = true
                return nil
            }
        }
        for _, f := range boolean_falses {
            if value == f {
                *(bool_map[uname]) = false
                return nil
            }
        }
        err_str := fmt.Sprintf("\"%s\" not a recognizable boolean", value)
        if verbose {
            fmt.Fprintf(os.Stderr, "%s\n", err_str)
        }
        return errors.New(err_str)
        
    } else {
        err_str := fmt.Sprintf("some logical error has led us here: %s=%s",
                               uname, value)
        if verbose {
            fmt.Fprintf(os.Stderr, "%s\n", err_str)
        }
        return errors.New(err_str)
    }
}

// Configure() reads a configuration file, setting the values of any
// configured variables to those found in the configuration file.
// The files argument is a slice of paths to possible configuration
// files; Configure() seeks them in order and processes the first one
// it finds. The verbose argument controls whether processing errors
// are written to stdout.
//
func Configure(files []string, verbose bool) error {   
    var cfg_file string
    for _, fname := range files {
        if _, err := os.Stat(fname); err == nil {
            cfg_file = fname
            break
        }
    }
    if cfg_file == "" {
        return errors.New("no configuration files found")
    }
    
    file, err := os.Open(cfg_file)
    if err != nil {
        err_str := fmt.Sprintf("error opening \"%s\"", cfg_file)
        if verbose {
            fmt.Fprintf(os.Stderr, "%s\n", err_str)
        }
        return errors.New(err_str)
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if comment_re.MatchString(line) {
            continue
        } else if !nonblank_re.MatchString(line) {
            continue
        }
        
        matches := option_re.FindStringSubmatch(line)
        if matches == nil {
            if verbose {
                fmt.Fprintf(os.Stderr, "ignoring malformed line: \"%s\"\n", line)
            }
        } else {
            setOption(matches[1], matches[2], verbose)
        }
    }
    
    return nil
}

func init() {
    comment_re = regexp.MustCompile(`^\s*#`)
    nonblank_re = regexp.MustCompile(`\S`)
    option_re = regexp.MustCompile(`^\s*([^:=]+)=(.*)$`)
    int_token = regexp.MustCompile(`-?\d+`)
    uint_token = regexp.MustCompile(`\d+`)
    float_token = regexp.MustCompile(`-?[0-9.]+`)
    ufloat_token = regexp.MustCompile(`[0-9.]+`)
    
    Reset()
}
