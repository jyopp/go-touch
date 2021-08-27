package touch_test

import (
	"image"
	"testing"

	touch "github.com/jyopp/go-touch"
)

func TestRectOperations(t *testing.T) {
	r1, r2 := image.Rect(0, 0, 10, 10), image.Rect(5, 0, 15, 10)
	t.Run("Winnow Subtracts from Rects", func(t *testing.T) {
		winnowed := touch.Winnow(r1, r2)
		if expect := image.Rect(0, 0, 5, 10); winnowed != expect {
			t.Logf("Winnow did not remove overlap: %v != %v", winnowed, expect)
		}
	})
	t.Run("Disjoin Separates Rects", func(t *testing.T) {
		touch.Disjoin(&r1, &r2)
		if union := r1.Union(r2); r1.Dx()+r2.Dx() != union.Dx() {
			t.Logf("Widths do not add to union width: (%d + %d) != %d", r1.Dx(), r2.Dx(), union.Dx())
			t.Fail()
		}
		if r1.Overlaps(r2) {
			t.Logf("Rects should be disjoint, but they overlap: %v, %v", r1, r2)
			t.Fail()
		}
	})
	t.Run("Merge Joins Adjacent Rects", func(t *testing.T) {
		merged := touch.Merge(r1, r2)
		if !r1.In(merged) || !r2.In(merged) {
			t.Logf("Merged rect does not cover inputs: %v != %v + %v", merged, r1, r2)
			t.Fail()
		}
	})

}

func naiveWinnow(r1, r2 image.Rectangle) image.Rectangle {
	r2 = r1.Intersect(r2)
	// Check each edge of bounds and attempt to trim it.
	if r2.Min == r1.Min {
		if r2.Max.Y == r1.Max.Y {
			// Cut pixels from left
			r1.Min.X = r2.Max.X
		} else if r2.Max.X == r1.Max.X {
			// Cut pixels from top
			r1.Min.Y = r2.Max.Y
		}
	} else if r2.Max == r1.Max {
		if r2.Min.Y == r1.Min.Y {
			// Cut pixels from right
			r1.Max.X = r2.Min.X
		} else if r2.Min.X == r1.Min.X {
			// Cut pixels from bottom
			r1.Max.Y = r2.Min.Y
		}
	}
	if r1.Min.X >= r1.Max.X || r1.Min.Y >= r1.Max.Y {
		return image.Rectangle{}
	}
	return r1
}

func BenchmarkRectOperations(b *testing.B) {
	// Create a slice of 256 rectangles with complete coverage over 0 ≤ x,y,w,h ≤ 3
	rects := make([]image.Rectangle, 256)
	for i := 0; i < 0xFF; i++ {
		rects[i].Min = image.Point{X: (i >> 6) & 0x3, Y: (i >> 4) & 0x3}
		rects[i].Max = rects[i].Min.Add(image.Point{X: (i >> 2) & 0x3, Y: i & 0x3})
	}

	b.Run("Winnow (Intersect-Based)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for testIdx := b.N; pb.Next(); testIdx++ {
				for i := 0; i < len(rects); i++ {
					for j := 0; j < len(rects); j++ {
						naiveWinnow(rects[i], rects[j]).Empty()
					}
				}
			}
		})
	})

	b.Run("Winnow (Custom)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for testIdx := b.N; pb.Next(); testIdx++ {
				for i := 0; i < len(rects); i++ {
					for j := 0; j < len(rects); j++ {
						touch.Winnow(rects[i], rects[j]).Empty()
					}
				}
			}
		})
	})

}
