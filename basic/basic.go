// Package basic - define several basic insights
package basic

import "golang.org/x/net/context"

// Insighter -- insight based on entry url
type Insighter interface {
	Insight(ctx context.Context)
}
