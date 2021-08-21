package fbui

import (
	"image"
	"sort"
)

type RegionList struct {
	Rects []image.Rectangle
}

// SetDirty expands or appends a dirty rect to include all pixels in rect.
func (rl *RegionList) AddRect(rect image.Rectangle) {
	for idx, oldRect := range rl.Rects {
		if rl.shouldMerge(rect, oldRect) {
			rl.Rects[idx] = oldRect.Union(rect)
			return
		}
	}
	rl.Rects = append(rl.Rects, rect)
}

// TODO: This can be made much more complex, returning an array-of-rects
// For example if r1.Intersect(r2).(Min,Max).Y == r2.(Min,Max).Y, the
// intersecting part of r2 should be removed and the exclusion returned.
func (rl *RegionList) shouldMerge(r1, r2 image.Rectangle) bool {
	if !r1.Overlaps(r2) {
		return false
	} else if r1.Min.Y <= r2.Min.Y {
		// r1.Min is above or at r2.Min
		return r2.Max.Y <= r1.Min.Y
	} else {
		return r2.Max.Y >= r1.Max.Y
	}
}

// Dequeue performs a best-effort to coalesce and remove overlapping rectangles
// by covering them with fewer, larger rectangles.
// The returned list is guaranteed not to contain any overlapping rectangles.
// Returns the coalesced rectangles and clears the region list.
func (rl *RegionList) Dequeue() []image.Rectangle {
	rects := rl.Rects
	if len(rects) == 0 {
		return nil
	}
	// Sort rects by their Min.Y in ascending order
	sort.Slice(rects, func(i, j int) bool {
		return rects[i].Min.Y < rects[j].Min.Y
	})
	// Reduce all overlapping rects.
	// This is a heuristic, vulnerable to some worst-case patterns.
	var newLength = len(rects)
	rect1 := rects[0]
	for idx := 1; idx < len(rects); idx++ {
		rect2 := rects[idx]
		if rect1.Overlaps(rect2) {
			rect1 = rect1.Union(rect2)
			rects[idx] = rect1
			rects[idx-1] = image.Rectangle{}
			newLength--
		} else {
			rect1 = rects[idx]
		}
	}
	// Ensure all the empty rects are moved to the end
	sort.Slice(rects, func(i, j int) bool {
		rI, rJ := rects[i], rects[j]
		if rI.Empty() {
			return false
		} else if rJ.Empty() {
			return true
		}
		return rects[i].Min.Y < rects[j].Min.Y
	})

	// Clear all rects in the set.
	rl.Rects = rl.Rects[:0]
	return rects[:newLength]
}
