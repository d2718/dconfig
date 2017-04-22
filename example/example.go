// example.go
//
// A demonstration of the use of package dconfig.
//
package main

import( "fmt"; "os"
        "github.com/d2718/dconfig"
)

var(
    number       int = 4
    pl_noun   string = "Go programmers"
    t_or_f      bool = true
    verb      string = "swallow"
    amt      float64 = 12.6
    mass_noun string = "kilograms of seawater"
)

func init() {
    
    // Suggested practice is to call dconfig.Reset(), then immediately
    // configure all options and call dconfig.Configure() without doing
    // anything else in between.
    
    dconfig.Reset()

    // syntax is
    // dconfig.AddXxx(OPTION_NAME, default_value, interpreting_options)
    //
    // See the first few dozen lines of dconfig.go for details about the
    // various interpreting_options.

    dconfig.AddInt(&number,       "integer_value",        dconfig.NONE)
    dconfig.AddString(&pl_noun,   "plural_noun",          dconfig.STRIP)
    dconfig.AddBool(&t_or_f,      "true_or_false")
    dconfig.AddString(&verb,      "verb",                 dconfig.STRIP)
    dconfig.AddFloat(&amt,        "real_valued_quantity", dconfig.NONE)
    dconfig.AddString(&mass_noun, "mass_noun",            dconfig.STRIP)
    
    err := dconfig.Configure([]string{"example.conf"}, true)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading configuration file: %s\n", err)
    }
}

func main() {
    var fmt_str string
    
    if t_or_f {
        fmt_str = "%d %s did indeed %s %f %s.\n"
    } else {
        fmt_str = "%d %s didn't %s %f %s.\n"
    }
    
    fmt.Printf(fmt_str, number, pl_noun, verb, amt, mass_noun)
}
