package matrix

import (
	"math"
	"testing"
)

const tol = 1e-9

func assertReconstructs(t *testing.T, a Matrix) {
	t.Helper()
	q, r, err := a.QR()
	if err != nil {
		t.Fatalf("QR returned error: %v", err)
	}
	assertProductEquals(t, q, r, a)
	assertOrthonormalColumns(t, q)
	assertUpperTriangular(t, r)
}

func assertProductEquals(t *testing.T, q, r, want Matrix) {
	t.Helper()
	for i := range want {
		for j := range want[i] {
			var sum float64
			for k := range r {
				sum += q[i][k] * r[k][j]
			}
			if math.Abs(sum-want[i][j]) > tol {
				t.Errorf("Q*R[%d][%d] = %g, want %g", i, j, sum, want[i][j])
			}
		}
	}
}

func assertOrthonormalColumns(t *testing.T, q Matrix) {
	t.Helper()
	n := len(q[0])
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			var dot float64
			for k := range q {
				dot += q[k][i] * q[k][j]
			}
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(dot-want) > tol {
				t.Errorf("Q^T*Q[%d][%d] = %g, want %g", i, j, dot, want)
			}
		}
	}
}

func assertUpperTriangular(t *testing.T, r Matrix) {
	t.Helper()
	for i := 1; i < len(r); i++ {
		for j := 0; j < i; j++ {
			if math.Abs(r[i][j]) > tol {
				t.Errorf("R[%d][%d] = %g, want 0 (upper triangular)", i, j, r[i][j])
			}
		}
	}
}

func TestQRSquare(t *testing.T) {
	assertReconstructs(t, Matrix{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}})
}

func TestQRRectangular(t *testing.T) {
	assertReconstructs(t, Matrix{{1, 2}, {3, 4}, {5, 6}, {7, 8}})
}

func TestQRSingleColumn(t *testing.T) {
	assertReconstructs(t, Matrix{{3}, {4}})
}

func TestQRIdentityIsItsOwnQ(t *testing.T) {
	q, r, err := Matrix{{1, 0}, {0, 1}}.QR()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q[0][0] != 1 || q[1][1] != 1 || r[0][0] != 1 || r[1][1] != 1 {
		t.Errorf("identity should factor as Q=I, R=I; got Q=%v R=%v", q, r)
	}
}

func TestQRErrors(t *testing.T) {
	cases := []struct {
		name  string
		input Matrix
		want  error
	}{
		{"empty", Matrix{}, ErrEmpty},
		{"empty row", Matrix{{}}, ErrEmpty},
		{"ragged rows", Matrix{{1, 2}, {3}}, ErrRagged},
		{"more columns than rows", Matrix{{1, 2, 3}}, ErrTooWide},
		{"dependent columns", Matrix{{1, 2}, {2, 4}}, ErrRankDeficit},
		{"null column", Matrix{{0, 1}, {0, 2}}, ErrRankDeficit},
		{"infinite value", Matrix{{1, 2}, {3, math.Inf(1)}}, ErrNotFinite},
		{"NaN value", Matrix{{1, 2}, {3, math.NaN()}}, ErrNotFinite},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := c.input.QR(); err != c.want {
				t.Errorf("got %v, want %v", err, c.want)
			}
		})
	}
}
