# Published packing instances

Test data only. Not part of the library.

| File | Source |
|---|---|
| `br1.txt` … `br7.txt` | Bischoff & Ratcliff container loading instances (classes 1–7), OR-Library |
| `loh-nee.txt` | Loh & Nee container loading instances |

Format, per file:

```
<number of problems>
<problem id> <seed>
<container width> <container length> <container depth>
<number of item types>
<type id> <w> <wVertical> <l> <lVertical> <d> <dVertical> <quantity>
...            (one line per item type)
```

`wVertical` / `lVertical` / `dVertical` are `0` or `1` and say whether that edge of
the item may point upwards, i.e. which orientations the instance allows.

Across `br1`..`br7` these flags fall out as follows, by item type and by quantity:

| allowed vertical edges | types | quantity |
|---|---|---|
| one | 1245 (17.1%) | 15799 (16.7%) |
| two | 2669 (36.6%) | 36545 (38.5%) |
| three | 3386 (46.4%) | 42547 (44.8%) |

No type allows none. The two-edge class is why the loader passes an explicit set
of permitted vertical axes rather than one of the three named rotation modes.

Utilisation figures recorded before the flags were applied were measured under
looser conditions than the instances state, and are not comparable with the
current ones.

References:

- E. E. Bischoff, M. S. W. Ratcliff, "Issues in the development of approaches to
  container loading", Omega 23(4), 1995.
- W. T. Loh, A. Y. C. Nee, "A packing algorithm for hexahedral boxes", Proceedings
  of the Industrial Automation Conference, 1992.

Files copied from the test suite of `dvdoug/BoxPacker` (MIT).

These files are read only by tests behind the `quality` build tag:

```sh
make test-quality   # pack every instance, compare against expected-utilisation.csv
make bench-quality  # fill rate and items left behind, per strategy
make baseline       # rewrite expected-utilisation.csv from the current run
```

An ordinary `go test ./...` does not compile those tests and never opens these files.

The first line of the baseline says what the readings were taken under: how many
instances of each class, which placement merit, and whether the free space
corners were offered. A run under other settings refuses to compare rather than
reading a number that means something else. Every rewrite is also appended to
`history.csv`, so a regression against a run three changes ago is as visible as
one against the last.

The baseline is a floor, not a record of the best ever seen: the test fails when a
change packs less, and is rewritten when a change packs more. It was last rewritten
after placement scoring started counting contact area, which lifted the mean over
br1-br7 from 0.8149 to 0.8534, and from 0.7437 to 0.7879 with full base support
enforced. Earlier numbers in the file's history were measured with looser
orientation flags and are not comparable with the current ones.

## What the published results say

The figures a reader should hold this library against, so a number here can be
read as good or bad rather than merely large:

| approach | BR1-7 mean fill | support |
|---|---|---|
| Bischoff & Ratcliff 1995, layer heuristic | 83.4% | full |
| Gehring & Bortfeldt 1997, genetic algorithm | 88.3% | full |
| Bortfeldt & Gehring 2001, tabu search | 91.3% | full |
| Parreño et al. 2008, maximal spaces with GRASP | 92.9-93.8% | none |
| Fanslau & Bortfeldt 2010, block building with tree search | 95.0% | none |
| Fanslau & Bortfeldt 2010, same, packing blocks | 94.2% | full |
| Araya & Riff 2014, beam search | ~96% | none |

References for those:

- E. E. Bischoff, M. S. W. Ratcliff, Omega 23(4), 1995.
- A. Bortfeldt, H. Gehring, "A hybrid genetic algorithm for the container
  loading problem", European Journal of Operational Research 131, 2001.
- F. Parreño, R. Alvarez-Valdes, J. F. Oliveira, J. M. Tamarit, "A maximal-space
  algorithm for the container loading problem", INFORMS Journal on Computing
  20(3), 2008.
- T. Fanslau, A. Bortfeldt, "A tree search algorithm for solving the container
  loading problem", INFORMS Journal on Computing 22(2), 2010.
- I. Araya, M.-C. Riff, "A beam search approach to the container loading
  problem", Computers & Operations Research 43, 2014.

| this library, default rule | 85.5% | none |
| this library, default rule | 79.2% | full |
| this library, `NewSearch(512, 3)` | 87.8% | none |
| this library, `NewSearch(512, 3)` | 85.9% | full |
| this library, `NewSearch(128, 3)` with free space corners | 88.8% | none |

The no-support and full-support columns are separate result tables in the
literature and are not comparable with each other. This library reports both.
Its own rows are thirty instances a class from `TestMeasure_Search`, not the
hundred each the published figures are averaged over, and are stated to be read
against the table above rather than in place of it.
