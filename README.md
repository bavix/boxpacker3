# boxpacker3

3D and 2D bin packing library for Go.

## Usage

```go
import (
  "context"
  "github.com/bavix/boxpacker3/v2"
)

packer := boxpacker3.NewPacker()

boxes := []*boxpacker3.Box{
  boxpacker3.NewBox("box std", 530, 380, 265, 20000),
}
items := []*boxpacker3.Item{
  boxpacker3.NewItem("product 1", 100, 100, 5, 2690),
  boxpacker3.NewItem("product 2", 100, 5, 100, 2690),
}

result, err := packer.Pack(context.Background(), boxes, items)
```

## Result

`Result.Boxes` lists used boxes, each with `PackedItem` values. `Result.Unpacked` lists items that did not fit, with a `Reason`.

## Algorithms

- **Greedy** — first fit, best fit, worst fit, next fit, fullest box
- **Search** — beam search with block-based expansion
- **Portfolio** — runs several algorithms in parallel, keeps the best

```go
packer := boxpacker3.NewPacker(
  boxpacker3.WithAlgorithm(boxpacker3.NewSearch(128, 3)),
)
```

## Constraints

- Weight limits, class separation, group co-location
- Placement rules, admission rules, pair rules
- Support ratio, load bearing, max load on top

## Goals

- `FewestBoxes`, `MostItems`, `LeastVolume`, `HighestFill`, `BalancedWeight`
- Custom goals via `Lexicographic`, `Weighted`, `CostGoal`

## Finishing

- `BalanceWeight` — even weight across boxes
- `Consolidate` — free the least-full box
- `Rescue` — re-offer unpacked items
- `Rehome` — move contents to smaller boxes
