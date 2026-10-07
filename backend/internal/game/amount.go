package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"regexp"
)

// Amount has no mutable exported storage. Every arithmetic result owns its value.
// Its zero value is the exact number zero.
type Amount struct{ value string }

var integerText = regexp.MustCompile(`^(0|-[1-9][0-9]*|[1-9][0-9]*)$`)

func NewAmount(n, d string) (Amount, error) {
	if !integerText.MatchString(n) || !integerText.MatchString(d) {
		return Amount{}, errors.New("invalid rational integer")
	}
	nn, _ := new(big.Int).SetString(n, 10)
	dd, _ := new(big.Int).SetString(d, 10)
	if dd.Sign() <= 0 {
		return Amount{}, errors.New("denominator must be positive")
	}
	return amountRat(new(big.Rat).SetFrac(nn, dd)), nil
}
func IntAmount(n int64) Amount { return amountRat(new(big.Rat).SetInt64(n)) }
func amountRat(r *big.Rat) Amount {
	if r.Sign() == 0 {
		return Amount{}
	}
	return Amount{r.RatString()}
}
func (a Amount) rat() *big.Rat { r, _ := new(big.Rat).SetString(a.String()); return r }
func (a Amount) String() string {
	if a.value == "" {
		return "0"
	}
	return a.value
}
func (a Amount) Add(b Amount) Amount { return amountRat(new(big.Rat).Add(a.rat(), b.rat())) }
func (a Amount) Sub(b Amount) Amount { return amountRat(new(big.Rat).Sub(a.rat(), b.rat())) }
func (a Amount) Mul(b Amount) Amount { return amountRat(new(big.Rat).Mul(a.rat(), b.rat())) }
func (a Amount) Quo(b Amount) (Amount, error) {
	if b.Sign() == 0 {
		return Amount{}, errors.New("division by zero")
	}
	return amountRat(new(big.Rat).Quo(a.rat(), b.rat())), nil
}
func (a Amount) Cmp(b Amount) int { return a.rat().Cmp(b.rat()) }
func (a Amount) Sign() int        { return a.rat().Sign() }
func (a Amount) Neg() Amount      { return amountRat(new(big.Rat).Neg(a.rat())) }
func (a Amount) MarshalJSON() ([]byte, error) {
	r := a.rat()
	return json.Marshal(struct {
		Numerator   string `json:"numerator"`
		Denominator string `json:"denominator"`
	}{r.Num().String(), r.Denom().String()})
}
func (a *Amount) UnmarshalJSON(b []byte) error {
	// A token walk prevents duplicate and case-folded field names accepted by encoding/json.
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return errors.New("rational object required")
	}
	vals := map[string]string{}
	for dec.More() {
		k, e := dec.Token()
		if e != nil {
			return e
		}
		key, ok := k.(string)
		if !ok || (key != "numerator" && key != "denominator") {
			return errors.New("unknown rational field")
		}
		if _, ok := vals[key]; ok {
			return errors.New("duplicate rational field")
		}
		var s string
		if e = dec.Decode(&s); e != nil {
			return e
		}
		vals[key] = s
	}
	if _, err = dec.Token(); err != nil {
		return err
	}
	if _, err = dec.Token(); err != io.EOF {
		return errors.New("trailing rational data")
	}
	v, err := NewAmount(vals["numerator"], vals["denominator"])
	if err != nil {
		return err
	}
	r := v.rat()
	if r.Num().String() != vals["numerator"] || r.Denom().String() != vals["denominator"] {
		return errors.New("noncanonical rational")
	}
	*a = v
	return nil
}

func (a Amount) IsZero() bool { return a.Sign() == 0 }
