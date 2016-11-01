package main

import (
    "math"
    "github.com/Knetic/govaluate"
)

var functions = map[string]govaluate.ExpressionFunction {
    // Trigonometrics
    "sin": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Sin(args[0].(float64))), nil
    },
    "cos": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Cos(args[0].(float64))), nil
    },
    "tan": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Tan(args[0].(float64))), nil
    },
    "sinh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Sinh(args[0].(float64))), nil
    },
    "cosh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Cosh(args[0].(float64))), nil
    },
    "tanh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Tanh(args[0].(float64))), nil
    },
    "arcsin": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Asin(args[0].(float64))), nil
    },
    "arccos": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Acos(args[0].(float64))), nil
    },
    "arctan": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Atan(args[0].(float64))), nil
    },
    "arcsinh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Asinh(args[0].(float64))), nil
    },
    "arccosh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Acosh(args[0].(float64))), nil
    },
    "arctanh": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Atanh(args[0].(float64))), nil
    },
    "hypot": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Hypot(args[0].(float64), args[1].(float64))), nil
    },
    // Roots
    "sqrt": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Sqrt(args[0].(float64))), nil
    },
    "cbrt": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Cbrt(args[0].(float64))), nil
    },
    // Logarithms
    "lb": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Log2(args[0].(float64))), nil
    },
    "ln": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Log(args[0].(float64))), nil
    },
    "lg": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Log10(args[0].(float64))), nil
    },
    // Exponentials
    "pow10": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Pow10(args[0].(int))), nil
    },
    "pow": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Pow(args[0].(float64), args[1].(float64))), nil
    },
    // Others
    "abs": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Abs(args[0].(float64))), nil
    },
    "ceil": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Ceil(args[0].(float64))), nil
    },
    "floor": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Floor(args[0].(float64))), nil
    },
    "mod": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Mod(args[0].(float64), args[1].(float64))), nil
    },
    "max": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Max(args[0].(float64), args[1].(float64))), nil
    },
    "min": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Min(args[0].(float64), args[1].(float64))), nil
    },
    "remainder": func(args ...interface{}) (interface{}, error) {
        return (float64)(math.Remainder(args[0].(float64), args[1].(float64))), nil
    },
}
