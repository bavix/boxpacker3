package boxpacker3

// Item represents an item that can be packed into a box.
type Item struct {
	id     string
	width  float64
	height float64
	depth  float64
	weight float64
	volume float64

	maxLength    float64
	rotationType RotationType
	position     Pivot
}

// itemSlice is a slice of items.
type itemSlice []*Item

// Len returns the length of the slice.
func (it itemSlice) Len() int {
	return len(it)
}

// Less returns true if the volume of the item at index i is less than the volume of the item at index j.
func (it itemSlice) Less(i, j int) bool {
	return it[i].volume < it[j].volume
}

// Swap swaps the items at index i and j.
func (it itemSlice) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

// NewItem creates a new item with the given parameters.
func NewItem(id string, w, h, d, wg float64) *Item {
	//nolint:exhaustruct
	return &Item{
		id:        id,
		width:     w,
		height:    h,
		depth:     d,
		weight:    wg,
		volume:    w * h * d,
		maxLength: max(w, h, d),
	}
}

// GetID returns the id of the item.
func (i *Item) GetID() string {
	return i.id
}

// GetWidth returns the width of the item.
func (i *Item) GetWidth() float64 {
	return i.width
}

// GetHeight returns the height of the item.
func (i *Item) GetHeight() float64 {
	return i.height
}

// GetDepth returns the depth of the item.
func (i *Item) GetDepth() float64 {
	return i.depth
}

// GetVolume returns the volume of the item.
func (i *Item) GetVolume() float64 {
	return i.volume
}

// GetWeight returns the weight of the item.
func (i *Item) GetWeight() float64 {
	return i.weight
}

// GetPosition returns the position of the item.
func (i *Item) GetPosition() Pivot {
	return i.position
}

// GetDimension returns the dimensions of the item based on its current rotation type.
//
// The dimensions are returned as a Dimension struct.
//
// Returns:
//   - Dimension: The dimensions of the item.
func (i *Item) GetDimension() Dimension {
	// Get the dimensions based on the rotation type of the item.
	switch i.rotationType {
	case RotationTypeWhd: // Width, Height, Depth
		return Dimension{i.GetWidth(), i.GetHeight(), i.GetDepth()}
	case RotationTypeHwd: // Height, Width, Depth
		return Dimension{i.GetHeight(), i.GetWidth(), i.GetDepth()}
	case RotationTypeHdw: // Height, Depth, Width
		return Dimension{i.GetHeight(), i.GetDepth(), i.GetWidth()}
	case RotationTypeDhw: // Depth, Height, Width
		return Dimension{i.GetDepth(), i.GetHeight(), i.GetWidth()}
	case RotationTypeDwh: // Depth, Width, Height
		return Dimension{i.GetDepth(), i.GetWidth(), i.GetHeight()}
	case RotationTypeWdh: // Width, Depth, Height
		return Dimension{i.GetWidth(), i.GetDepth(), i.GetHeight()}
	default: // RotationTypeWhd
		return Dimension{i.GetWidth(), i.GetHeight(), i.GetDepth()}
	}
}

// Intersect tests for intersections between two items.
//
// It checks for intersections between the current item and the given item.
// It does this by getting the dimensions of the current item and the given item
// and then calling the intersect method with the appropriate parameters.
//
// Parameters:
//   - it: The item to check for intersections with.
//
// Returns:
//   - bool: True if the items intersect, false otherwise.
func (i *Item) Intersect(it *Item) bool {
	// Get the dimensions of the current item and the given item.
	d1 := i.GetDimension()
	d2 := it.GetDimension()

	// Check for intersections in the x and y axes of the two items.
	return i.intersect(d1, d2, it, WidthAxis, HeightAxis) &&
		i.intersect(d1, d2, it, HeightAxis, DepthAxis) &&
		i.intersect(d1, d2, it, WidthAxis, DepthAxis)
}

// intersect Checks if two rectangles intersect from the x and y axes of elements i1 and i2.
//
// This function takes the following parameters:
//   - d1: The dimensions of the first rectangle.
//   - d2: The dimensions of the second rectangle.
//   - it: The second rectangle.
//   - x: The x axis for the intersection check.
//   - y: The y axis for the intersection check.
//
// It returns true if the rectangles intersect, false otherwise.
func (i *Item) intersect(d1, d2 Dimension, it *Item, x, y Axis) bool {
	// Calculate the center points of the two rectangles.
	cx1 := i.position[x] + d1[x]/2  //nolint:mnd
	cy1 := i.position[y] + d1[y]/2  //nolint:mnd
	cx2 := it.position[x] + d2[x]/2 //nolint:mnd
	cy2 := it.position[y] + d2[y]/2 //nolint:mnd

	// Calculate the intersection points on the x and y axes.
	ix := max(cx1, cx2) - min(cx1, cx2)
	iy := max(cy1, cy2) - min(cy1, cy2)

	// Check if the rectangles intersect.
	return ix < (d1[x]+d2[x])/2 && iy < (d1[y]+d2[y])/2
}
