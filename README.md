# boxpacker3

A 3D bin packing library in golang.

### Usage

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

---
Supported by

[![Supported by JetBrains](https://cdn.rawgit.com/bavix/development-through/46475b4b/jetbrains.svg)](https://www.jetbrains.com/)
