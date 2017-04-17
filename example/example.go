// example.go
//
// A demonstration of the use of package dconfig.
//
package main

import( "fmt"; "os"
        "github.com/d2718/dconfig"
)

var(
    number int = 4
    pl_noun string = "Go programmers"
    t_or_f bool = true
    verb string = "swallow"
    amt float64 = 12.6
    mass_noun string = "kilograms of seawater"
)

func init() {

    // syntax is
    // dconfig.AddXxx(OPTION_NAME, default_value, interpreting_options)
    //
    // See the first few dozen lines of dconfig.go for details about the
    // various interpreting_options.

    dconfig.AddInt("integer_value", number, dconfig.NONE)
    dconfig.AddString("plural_noun", pl_noun, dconfig.STRIP)
    dconfig.AddBool("true_or_false", t_or_f)
    dconfig.AddString("verb", verb, dconfig.STRIP)
    dconfig.AddFloat("real_valued_quantity", amt, dconfig.NONE)
    dconfig.AddString("mass_noun", mass_noun, dconfig.STRIP)
    
    err := dconfig.Configure([]string{"example.conf"}, true)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading configuration file: %s\n", err)
    }
    
    // dconfig.GetXxx() will return a zero value and an error if the
    // given option isn't present.

    number,    _ = dconfig.GetInt("integer_value")
    pl_noun,   _ = dconfig.GetString("plural_noun")
    t_or_f,    _ = dconfig.GetBool("true_or_false")
    verb,      _ = dconfig.GetString("verb")
    amt,       _ = dconfig.GetFloat("real_valued_quantity")
    mass_noun, _ = dconfig.GetString("mass_noun")
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
