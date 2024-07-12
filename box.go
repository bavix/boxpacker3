package boxpacker3

// Box represents a box that can hold items.
//
// It has fields for the box's dimensions and maximum weight.
// It also has fields for tracking the box's current items and their volume and weight.
type Box struct {
	// id is the box's unique identifier.
	id string

	// width is the box's width.
	width float64

	// height is the box's height.
	height float64

	// depth is the box's depth.
	depth float64

	// maxWeight is the maximum weight the box can hold.
	maxWeight float64

	// volume is the box's volume (width * height * depth).
	volume float64

	// items is a slice of items currently in the box.
	items []*Item

	// maxLength is the length of the box's longest side.
	maxLength float64

	// itemsVolume is the total volume of the items in the box.
	itemsVolume float64

	// itemsWeight is the total weight of the items in the box.
	itemsWeight float64
}

// boxSlice is a slice of boxes.
//
// It implements the sort.Interface by defining Len, Less and Swap methods.
type boxSlice []*Box

// Len returns the length of the boxSlice.
func (bs boxSlice) Len() int {
	return len(bs)
}

// Less compares two boxes by volume.
func (bs boxSlice) Less(i, j int) bool {
	return bs[i].volume < bs[j].volume
}

// Swap swaps two boxes in the boxSlice.
func (bs boxSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

// NewBox creates a new Box with the given id, dimensions, and maximum weight.
//
// Parameters:
// - id: a unique identifier for the box.
// - w: the width of the box.
// - h: the height of the box.
// - d: the depth of the box.
// - mw: the maximum weight the box can hold.
//
// Returns:
// - A pointer to the newly created Box.
func NewBox(id string, w, h, d, mw float64) *Box {
	//nolint:exhaustruct
	return &Box{
		id:        id,
		width:     w,
		height:    h,
		depth:     d,
		maxWeight: mw,
		maxLength: max(w, h, d),
		volume:    w * h * d,
		items:     make([]*Item, 0, 1),
	}
}

// GetID returns the unique identifier of the box.
func (b *Box) GetID() string {
	return b.id
}

// GetWidth returns the width of the box.
func (b *Box) GetWidth() float64 {
	return b.width
}

// GetHeight returns the height of the box.
func (b *Box) GetHeight() float64 {
	return b.height
}

// GetDepth returns the depth of the box.
func (b *Box) GetDepth() float64 {
	return b.depth
}

// GetVolume returns the volume of the box.
func (b *Box) GetVolume() float64 {
	return b.volume
}

// GetMaxWeight returns the maximum weight the box can hold.
func (b *Box) GetMaxWeight() float64 {
	return b.maxWeight
}

// GetItems returns a slice of pointers to the items currently in the box.
//
// The slice is a copy and not a reference to the original slice, so modifying
// the slice returned by this function will not affect the contents of the box.
func (b *Box) GetItems() []*Item {
	return append([]*Item(nil), b.items...)
}

// PutItem Attempts to place the given item at the specified anchor point within the box.
//
// Attempts to place the given item at the specified anchor point within the box.
//
// It tries to place the item at the given anchor point by iterating through each
// rotation type (Whd, Hwd, Hdw, Dhw, Dwh, Wdh) and checks if the item can be
// placed within the box without intersecting with any of the other items in the box.
// If the item can be placed, it inserts the item into the box and returns true.
// If the item cannot be placed, it returns false.
//
// Parameters:
//   - item: The item to be placed in the box.
//   - p: The anchor point at which to attempt placing the item within the box.
//
// Returns:
//   - bool: True if the item was successfully placed within the box, false otherwise.
func (b *Box) PutItem(item *Item, p Pivot) bool {
	// Check if the item can fit in the box based on volume and weight quotas.
	if !b.canQuota(item) {
		return false
	}

	// Set the item's position to the anchor point.
	item.position = p

	// Iterate through each rotation type to find a suitable placement.
	for rt := RotationTypeWhd; rt <= RotationTypeWdh; rt++ {
		// Set the item's rotation type to the current rotation type.
		item.rotationType = rt

		// Get the dimensions of the item in its current rotation type.
		itemDimensions := item.GetDimension()

		// Check if the box has enough dimensions to accommodate the item.
		if b.width < p[WidthAxis]+itemDimensions[WidthAxis] ||
			b.height < p[HeightAxis]+itemDimensions[HeightAxis] ||
			b.depth < p[DepthAxis]+itemDimensions[DepthAxis] {
			continue
		}

		// Check if the item intersects with any other items in the box.
		if b.itemsIntersect(item) {
			continue
		}

		// Insert the item into the box and return true.
		b.insert(item)

		return true
	}

	// If no suitable placement is found, return false.
	return false
}

// itemsIntersect checks if any of the items in the box intersect with the given item.
// It iterates through each item in the box and calls the Intersect method on the item.
// If an intersection is found, it returns true.
// If no intersection is found, it returns false.
func (b *Box) itemsIntersect(item *Item) bool {
	for _, ib := range b.items {
		if ib.Intersect(item) {
			return true
		}
	}

	return false
}

// canQuota checks if the box can accommodate the given item based on both volume and weight quotas.
//
// It calls the canFitVolume and canFitWeight methods to check if the box has enough room for the
// item's volume and weight. If both conditions are true, it returns true. Otherwise, it returns false.
func (b *Box) canQuota(item *Item) bool {
	return b.canFitVolume(item) && b.canFitWeight(item)
}

// canFitVolume checks if the box can accommodate the given item based on volume.
//
// It compares the sum of the item's volume and the current volume of items in the box
// to the box's total volume. If the sum is less than or equal to the box's total volume,
// it returns true. Otherwise, it returns false.
func (b *Box) canFitVolume(item *Item) bool {
	return b.itemsVolume+item.volume <= b.volume
}

// canFitWeight checks if the box can accommodate the given item based on weight.
//
// It compares the sum of the item's weight and the current weight of items in the box
// to the box's maximum weight. If the sum is less than or equal to the box's maximum weight,
// it returns true. Otherwise, it returns false.
func (b *Box) canFitWeight(item *Item) bool {
	return b.itemsWeight+item.weight <= b.maxWeight
}

// insert inserts an item into the box and updates the total volume and weight.
//
// It appends the item to the box's items slice and adds the item's volume and weight to the
// box's total volume and weight.
func (b *Box) insert(item *Item) {
	b.items = append(b.items, item)
	b.itemsVolume += item.volume
	b.itemsWeight += item.weight
}

// Reset clears the box and resets the volume and weight.
//
// It removes all items from the box by slicing the items slice to an empty slice.
// It sets the total volume and weight of items in the box to 0.
func (b *Box) Reset() {
	b.items = b.items[:0]
	b.itemsVolume = 0
	b.itemsWeight = 0
}
