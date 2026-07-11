package elo

import "math"

// KFactor controls how much a single game can move a rating.
const KFactor = 32

// Score values for the white player's result.
const (
	Win  = 1.0
	Draw = 0.5
	Loss = 0.0
)

// NewRatings returns the updated ratings for two players given the first
// player's score (Win/Draw/Loss from playerA's perspective). Standard Elo
// with a fixed K-factor.
func NewRatings(ratingA, ratingB int32, scoreA float64) (int32, int32) {
	expectedA := 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
	expectedB := 1.0 - expectedA

	scoreB := 1.0 - scoreA

	newA := float64(ratingA) + KFactor*(scoreA-expectedA)
	newB := float64(ratingB) + KFactor*(scoreB-expectedB)

	return int32(math.Round(newA)), int32(math.Round(newB))
}
