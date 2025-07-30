package fuzzy

import "testing"

func TestTriangular(t *testing.T) {
	triangular := Triangular(-1, 0, 1)

	if g, e := triangular.Value(-2), 0.0; g != e {
		t.Errorf("triangular(-2): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(-1), 0.0; g != e {
		t.Errorf("triangular(-1): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(-0.5), 0.5; g != e {
		t.Errorf("triangular(-0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(0), 1.0; g != e {
		t.Errorf("triangular(0): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(0.5), 0.5; g != e {
		t.Errorf("triangular(0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(1), 0.0; g != e {
		t.Errorf("triangular(1): got '%v', expected '%v'", g, e)
	}

	if g, e := triangular.Value(2), 0.0; g != e {
		t.Errorf("triangular(2): got '%v', expected '%v'", g, e)
	}
}

func TestInvertedTriangular(t *testing.T) {
	invertedTriangular := Inverted(Triangular(-1, 0, 1))

	if g, e := invertedTriangular.Value(-2), 1.0; g != e {
		t.Errorf("invertedTriangular(-2): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(-1), 1.0; g != e {
		t.Errorf("invertedTriangular(-1): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(-0.5), 0.5; g != e {
		t.Errorf("invertedTriangular(-0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(0), 0.0; g != e {
		t.Errorf("invertedTriangular(0): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(0.5), 0.5; g != e {
		t.Errorf("invertedTriangular(0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(1), 1.0; g != e {
		t.Errorf("invertedTriangular(1): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedTriangular.Value(2), 1.0; g != e {
		t.Errorf("invertedTriangular(2): got '%v', expected '%v'", g, e)
	}
}

func TestLinear(t *testing.T) {
	linear := Linear(0, 1)

	if g, e := linear.Value(-1), 0.0; g != e {
		t.Errorf("linear(-1): got '%v', expected '%v'", g, e)
	}

	if g, e := linear.Value(0), 0.0; g != e {
		t.Errorf("linear(0): got '%v', expected '%v'", g, e)
	}

	if g, e := linear.Value(0.5), 0.5; g != e {
		t.Errorf("linear(0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := linear.Value(0.25), 0.25; g != e {
		t.Errorf("linear(0.25): got '%v', expected '%v'", g, e)
	}

	if g, e := linear.Value(1), 1.0; g != e {
		t.Errorf("linear(1): got '%v', expected '%v'", g, e)
	}

	if g, e := linear.Value(2), 1.0; g != e {
		t.Errorf("linear(2): got '%v', expected '%v'", g, e)
	}
}

func TestInvertedLinear(t *testing.T) {
	invertedLinear := Inverted(Linear(0, 1))

	if g, e := invertedLinear.Value(-1), 1.0; g != e {
		t.Errorf("invertedLinear(-1): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedLinear.Value(0), 1.0; g != e {
		t.Errorf("invertedLinear(0): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedLinear.Value(0.25), 0.75; g != e {
		t.Errorf("invertedLinear(0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedLinear.Value(0.5), 0.5; g != e {
		t.Errorf("invertedLinear(0.5): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedLinear.Value(1), 0.0; g != e {
		t.Errorf("invertedLinear(1): got '%v', expected '%v'", g, e)
	}

	if g, e := invertedLinear.Value(2), 0.0; g != e {
		t.Errorf("invertedLinear(2): got '%v', expected '%v'", g, e)
	}
}

func TestTrapezoid(t *testing.T) {
	trapezoid := Trapezoid(10, 20, 30, 40)

	if g, e := trapezoid.Value(5), 0.0; g != e {
		t.Errorf("trapezoid(5): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(10), 0.0; g != e {
		t.Errorf("trapezoid(10): got '%v', expected '%v'", g, e)
	}

	if g, e := trapezoid.Value(15), 0.5; g != e {
		t.Errorf("trapezoid(15): got '%v', expected '%v'", g, e)
	}

	if g, e := trapezoid.Value(20), 1.0; g != e {
		t.Errorf("trapezoid(20): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(25), 1.0; g != e {
		t.Errorf("trapezoid(25): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(30), 1.0; g != e {
		t.Errorf("trapezoid(30): got '%v', expected '%v'", g, e)
	}

	if g, e := trapezoid.Value(35), 0.5; g != e {
		t.Errorf("trapezoid(35): got '%v', expected '%v'", g, e)
	}

	if g, e := trapezoid.Value(40), 0.0; g != e {
		t.Errorf("trapezoid(40): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(45), 0.0; g != e {
		t.Errorf("trapezoid(45): got '%v', expected '%v'", g, e)
	}
}

func TestTrapezoidAsTriangle(t *testing.T) {
	trapezoid := Trapezoid(10, 20, 20, 30) // Should behave like Triangular(10, 20, 30)

	if g, e := trapezoid.Value(15), 0.5; g != e {
		t.Errorf("trapezoidAsTriangle(15): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(20), 1.0; g != e {
		t.Errorf("trapezoidAsTriangle(20): got '%v', expected '%v'", g, e)
	}
	if g, e := trapezoid.Value(25), 0.5; g != e {
		t.Errorf("trapezoidAsTriangle(25): got '%v', expected '%v'", g, e)
	}
}
