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
		if merged := Merge(rect, oldRect); !merged.Empty() {
			rl.Rects[idx] = merged
			return
		}
	}
	rl.Rects = append(rl.Rects, rect)
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
	// Reduce all overlapping rects. Any overlap results in replacement with the Union.
	// This is a heuristic, vulnerable to some worst-case patterns.
	// Unnecessary areas may be added, but no area will be covered by multiple rects.
	var newLength = len(rects)
	for idx := 1; idx < len(rects); idx++ {
		// TODO: Something with Merge/Disjoin here, but with sensitivity to whether
		// fewer pixels can be drawn. Perhaps track different rects for drawing vs.
		// flushing to the buffer. As is, Disjoin would break the guarantee that
		// pixels cannot be covered twice, since we don't know which of the two
		// reduced rects (or even both) might overlap a future rect.
		if rects[idx].Overlaps(rects[idx-1]) {
			rects[idx] = rects[idx].Union(rects[idx-1])
			rects[idx-1] = image.Rectangle{}
			newLength--
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
