package fbui

import "image"

// Winnow returns a copy of rect after removing any edge that is completely covered
// by the rect to remove. If rect is completely covered, returns a zero-valued rectangle.
func Winnow(rect, remove image.Rectangle) image.Rectangle {
	// If 'remove' covers the entire width of rect, try to clip top / bottom
	if remove.Min.X <= rect.Min.X && remove.Max.X >= rect.Max.X {
		if remove.Min.Y <= rect.Min.Y && remove.Max.Y > rect.Min.Y {
			// Remove covered area from the top
			rect.Min.Y = remove.Max.Y
		} else if remove.Max.Y >= rect.Max.Y && remove.Min.Y < rect.Max.Y {
			// Remove covered area from the bottom
			rect.Max.Y = remove.Min.Y
		}
	}
	if remove.Min.Y <= rect.Min.Y && remove.Max.Y >= rect.Max.Y {
		if remove.Min.X <= rect.Min.X && remove.Max.X > rect.Min.X {
			// Remove covered area from the top
			rect.Min.X = remove.Max.X
		} else if remove.Max.X >= rect.Max.X && remove.Min.X < rect.Max.X {
			// Remove covered area from the bottom
			rect.Max.X = remove.Min.X
		}
	}
	if rect.Max.Y <= rect.Min.Y || rect.Max.X <= rect.Min.X {
		rect = image.Rectangle{}
	}
	return rect
}

// Disjoin makes an effort to replace two overlapping rectangles with two smaller,
// nonoverlapping rectangles. If the rectangles do not overlap, or if neither rectangle
// completely covers any edge of the other, the rectangles are not modified.
func Disjoin(r1, r2 *image.Rectangle) (modified bool) {
	if rect := Winnow(*r1, *r2); rect != *r1 {
		*r1 = rect
		modified = true
	} else if rect := Winnow(*r2, *r1); rect != *r2 {
		*r2 = rect
		modified = true
	}
	return
}

// Merge attempts to combine two rectangles that share an edge.
// If a single rectangle can be created that covers the exact same area, it is returned.
// If no such rectangle exists, returns a zero-valued rectangle.
func Merge(r1, r2 image.Rectangle) image.Rectangle {
	// Handle disqualifying checks as early as possible
	if r2.Min.Y > r1.Max.Y || r2.Max.Y < r1.Min.Y || r2.Min.X > r1.Max.X || r2.Max.X < r1.Max.X {
		return image.Rectangle{}
	}

	if r1.Min.X == r2.Min.X && r1.Max.X == r2.Max.X {
		// Rectangles cover the same pixels in width
		// These rectangles cannot be disjoint; see preconditions above
		// As such, we can definitely merge them.
		if r1.Min.Y > r2.Min.Y {
			r1.Min.Y = r2.Min.Y // Extend at top
		}
		if r1.Max.Y < r2.Max.Y {
			r1.Max.Y = r2.Max.Y // Extend at bottom
		}
		return r1
	} else if r1.Min.Y == r2.Min.Y && r1.Max.Y == r2.Max.Y {
		if r1.Min.Y > r2.Min.Y {
			r1.Min.Y = r2.Min.Y // Extend to left
		}
		if r1.Max.X < r2.Max.X {
			r1.Max.X = r2.Max.X // Extend to right
		}
		return r1
	} else if r2.In(r1) {
		return r1
	} else if r1.In(r2) {
		return r2
	}
	return image.Rectangle{}
}
