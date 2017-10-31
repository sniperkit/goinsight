// Package basic - define several basic insights
package basic

// Insighter -- insight based on entry url
type Insighter interface {
	Insight(url string)
}
