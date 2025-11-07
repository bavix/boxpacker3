# boxpacker3

A 3D and 2D bin packing library in golang.

## Usage

```golang
packer := boxpacker3.NewPacker()

boxes := []*boxpacker3.Box{
  boxpacker3.NewBox("box std", 530, 380, 265, 20000),
}
items := []*boxpacker3.Item{
  boxpacker3.NewItem("product 1", 100, 100, 5, 2690),
  boxpacker3.NewItem("product 2", 100, 5, 100, 2690),
  boxpacker3.NewItem("product 3", 5, 100, 100, 2690),
}

packResult := packer.Pack(boxes, items)
fmt.Println(packResult.Boxes) // boxes and items
fmt.Println(packResult.UnfitItems) // Items that didn't fit in boxes
```

## 2D Packing

The library also supports 2D packing for flat items like sheets, boards, or panels. Use `NewBox2D` and `NewItem2D` to create 2D boxes and items (depth is automatically set to 1):

```golang
packer := boxpacker3.NewPacker()

boxes := []*boxpacker3.Box{
  boxpacker3.NewBox2D("box 2d", 100, 50, 1000), // width=100, height=50, depth=1
}
items := []*boxpacker3.Item{
  boxpacker3.NewItem2D("sheet 1", 30, 20, 100), // width=30, height=20, depth=1
  boxpacker3.NewItem2D("sheet 2", 40, 25, 100),
}

packResult := packer.Pack(boxes, items)
```

2D boxes and items work seamlessly with all packing strategies and can be mixed with regular 3D boxes and items if needed.

## Packing Strategies

The library supports multiple packing strategies that can be selected when creating a packer:

### StrategyMinimizeBoxes (Default)
Minimizes the number of boxes used. Sorts items by volume in descending order (largest first) and uses First Fit algorithm. Best for minimizing box count.

### StrategyGreedy
Simple and fast strategy. Sorts items by volume in ascending order (smallest first) and uses First Fit algorithm. May use more boxes than optimal strategies.

### StrategyBestFit
For each item, finds the box with the smallest remaining space that can accommodate it. Minimizes wasted space but requires checking all boxes for each item.

### StrategyBestFitDecreasing
Items are sorted by volume in descending order, and for each item, finds the box with the smallest remaining space. Typically provides 2-5% better space utilization than First Fit Decreasing.

### StrategyNextFit
Items are placed in the current box if it fits, otherwise a new box is used. Simpler than First Fit but may use more boxes.

### StrategyWorstFit
For each item, finds the box with the largest remaining space that can accommodate it. Helps distribute items more evenly across boxes.

### StrategyAlmostWorstFit
Similar to Worst Fit, but excludes boxes that are too large (almost empty). Prevents items from being placed in boxes that are nearly empty.

## Example with Strategy

```golang
packer := boxpacker3.NewPacker(
  boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing),
)

packResult := packer.Pack(boxes, items)
```
