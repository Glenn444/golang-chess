package elo

import "testing"

func TestEqualRatingsWin(t *testing.T) {
	a, b := NewRatings(1200, 1200, Win)
	if a != 1216 || b != 1184 {
		t.Errorf("expected 1216/1184, got %d/%d", a, b)
	}
}

func TestEqualRatingsDraw(t *testing.T) {
	a, b := NewRatings(1200, 1200, Draw)
	if a != 1200 || b != 1200 {
		t.Errorf("draw between equals must not change ratings, got %d/%d", a, b)
	}
}

func TestUnderdogWinGainsMore(t *testing.T) {
	a, b := NewRatings(1200, 1400, Win)
	gain := a - 1200
	loss := int32(1400) - b
	if gain <= KFactor/2 {
		t.Errorf("underdog should gain more than half K, gained %d", gain)
	}
	if gain != loss {
		t.Errorf("zero-sum violated: gain %d, loss %d", gain, loss)
	}
}

func TestFavoriteWinGainsLittle(t *testing.T) {
	a, _ := NewRatings(1400, 1200, Win)
	if gain := a - 1400; gain >= KFactor/2 {
		t.Errorf("favorite should gain less than half K, gained %d", gain)
	}
}

func TestLoss(t *testing.T) {
	a, b := NewRatings(1200, 1200, Loss)
	if a != 1184 || b != 1216 {
		t.Errorf("expected 1184/1216, got %d/%d", a, b)
	}
}
