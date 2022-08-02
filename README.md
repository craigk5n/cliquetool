# About Cliquetool

Multi-purpose tool written in Go for finding cliques within a graph,
building dense groups from the cliques, and generating reports on the results.

This project does a number of things with an unidirected graph loaded
from a CSV file:
- Find the maximum clique from a set of links (aka edges) while
  it also finding all cliques
- Find most of the dense graphs by using
  the found cliques as the basis to merge groups of nodes into nearly
  complete graphs
- Analyze results to generate reports on the best candidate nodes
  to build each group into a larger (slightly less connected) group

## Terminology

In the source code, "node" and "link" are used rather than the more common
(at least in CS graph theory) vertex and edge.  The terms "group" is
used the code to refer to a graph.
A "group of nodes conected by links" is the same
as a "graph of vertices connected by edges"

## What is the maximum clique
A clique in a graph is a set of nodes, where every node
in this set is adjacent (connected) to every other in the set.
In other words, each node in this max clique set is connected to every other
node.
Finding the size of a maximum
clique in a given graph is one of the fundamental
[NP-hard](https://en.wikipedia.org/wiki/NP-hardness) problems.

## The Algorithm

I actually developed the algorithm for finding cliques on my own many
years ago before realing this was a long-studied area of Computer Science.
A coworker pointed out some of the research around graph theory and
I realized I reinvented something very similar to what others had
already done.

This is a
["branch and bound"](https://www.geeksforgeeks.org/branch-and-bound-algorithm/)
solution to the Maximum Clique.  It works well for sparsely connected data.
It runs pretty fast on some of the included test data, but it will slow to
a crawl on some of the highly-connected DIMACS data below (like many similar
algorithms).

## But Why?

Graph theory is a fascinating part of Computer Science.  I had originally written
this algorithm in C many years ago and wanted to try to implement it again in Go
(or at least something very similar)
as an exercise to learn Go.
(I no longer have the C implementation, but I think this is pretty close.)


## Data

Included in this repo are a few sample data sets.
The DIMACS data is what is traditionally used to evaluate
graph analysis tools.  However, as a college sports fan,
the basketball and football schedules are are more fun
challenge.

### NCAA Basketball Schedule for 2018-2019

This data set is the complete schedule of NCAA basketball
games for the 2018-19 season.  From this data set you should be able to
determine the conference membership since each team plays
every other team in the same conference at least once (typically twice).
And no team that is not a member of a conference plays all teams in
that conference.  This is a simple example where finding all the cliques
within the data will find all the conferences.
The data file is in the
[testdata](testdata/) subdirectory
as [2018_NCAAB.csv](testdata/2018_NCAAB.csv).

### NCAA Football Schedule for 2002

The original data can be found
[here](http://www.phys.utk.edu/sorensen/cfr/cfr/Output/2002/CF_2002_Ascii_Games.txt).
You can check the results against the actual conference membership
[here](http://www.phys.utk.edu/sorensen/cfr/cfr/Output/2002/CF_2002_Conferences.html).

#### Report Output
The full report can be found in the repo
[here](doc/2002_NCAA_report.md).  But here are some of the tables generated as
part of the report.  In this example, the `cliquetool` is used like below to
find all the cliques and then build the dense (not fully connected) groups:
```
./cliquetool -i testdata/2002_NCAAF.csv -minsize 5 -find -sort -rename \
  -writecsv 2002_NCAAF-cliques.csv -param testdata/2002_NCAAF-params.txt \
  -passes 3 -build -mergeparams 0.7 -merge -prune -prefix NCAAF_Conf_ \
  -sort -rename \
  -minsize 8 -writecsv 2002_NCAAF-build-groups.csv
```
And then the following command will generate the markdown report:
```
./cliquetool -i testdata/2002_NCAAF.csv \
  -loadgroupcsv 2002_NCAAF-build-groups.csv -verify doc/2002_NCAA_report.md
```

##### Group Size Summary
|      Group Size |       Count (=) |      Count (>=) |
|             --- |            ---: |            ---: |
|              12 |               1 |               1 |
|              11 |               4 |               5 |
|              10 |              12 |              17 |
|               9 |               9 |              26 |
|               8 |              13 |              39 |

##### Group Density Summary
|   Group Density |       Count (=) |      Count (>=) |
|             --- |            ---: |            ---: |
|            100% |              31 |              31 |
|      90.0-99.9% |               5 |              36 |
|      80.0-89.9% |               3 |              39 |
|      70.0-79.9% |               0 |              39 |
|      60.0-69.9% |               0 |              39 |
|      50.0-59.9% |               0 |              39 |
|      40.0-49.9% |               0 |              39 |
|      30.0-39.9% |               0 |              39 |
|      20.0-29.9% |               0 |              39 |
|      10.0-19.9% |               0 |              39 |
|      00.0-09.9% |               0 |              39 |

##### Node Group Membership Counts
|      No. Groups |       No. Nodes |
|            ---: |            ---: |
|       1+ Groups |    361 (51.28%) |
|       No Groups |    343 (48.72%) |
|               1 |    361 (51.28%) |
|               2 |       0 (0.00%) |

##### Group Statistics
|                      Name |     Size |  Group Links |  Total Links |   Density |
|                       --- |     ---: |         ---: |         ---: |      ---: |
|          NCAAF_Conf_00001 |       12 |          124 |          136 |     93.94 |
|          NCAAF_Conf_00002 |       11 |          100 |          135 |     90.91 |
|          NCAAF_Conf_00003 |       11 |          110 |          119 |    100.00 |
|          NCAAF_Conf_00004 |       11 |          100 |          113 |     90.91 |
|          NCAAF_Conf_00005 |       11 |          110 |          119 |    100.00 |
|          NCAAF_Conf_00006 |       10 |           80 |           99 |     88.89 |
|          NCAAF_Conf_00007 |       10 |           90 |          101 |    100.00 |
... and many more

and this last table takes a look at each group and shows which nodes were not connected
that were considered but discarded because they did not meet the criteria:

##### Summary For Group: NCAAF_Conf_00001
|                    NodeID |    Group Links |   Grp/Total Lnks | No. Groups |
|                       --- |           ---: |             ---: |       ---: |
|        Findlay University |  10/11  90.91% |    10/11  90.91% |          1
|            Wayne State MI |  10/11  90.91% |    10/11  90.91% |          1
|        Saginaw Valley Sta |  11/11 100.00% |    11/12  91.67% |          1
|              Ferris State |  11/11 100.00% |    11/11 100.00% |          1
|         Hillsdale College |  11/11 100.00% |    11/11 100.00% |          1
|        Northwood Universi |  11/11 100.00% |    11/11 100.00% |          1
|             Michigan Tech |  10/11  90.91% |    10/10 100.00% |          1
|        Ashland University |  10/11  90.91% |    10/11  90.91% |          1
|        Mercyhurst College |  10/11  90.91% |    10/11  90.91% |          1
|              Indianapolis |  10/11  90.91% |    10/11  90.91% |          1
|         Northern Michigan |  10/11  90.91% |    10/11  90.91% |          1
|        Grand Valley State |  10/11  90.91% |    10/15  66.67% |          1
|                    TOTALS | 124/132  93.94% |      AVE  91.98% |          - |
###### Potential group nodes not included
|                    NodeID |    Group Links |   Grp/Total Lnks | No. Groups |
|                       --- |           ---: |             ---: |       ---: |
|                Indiana PA |   3/12  25.00% |     3/13  23.08% |          0 |
|             Northern Iowa |   1/12   8.33% |     1/11   9.09% |          1 |
|         Edinboro Universi |   1/12   8.33% |     1/11   9.09% |          0 |
|         West Virginia Wes |   1/12   8.33% |     1/11   9.09% |          1 |
|         Saint Joseph's IN |   1/12   8.33% |     1/11   9.09% |          0 |

### DIMACS Benchmark Set

There is some useful data for finding the max clique in the following data sets:
- [DIMACS benchmark set](https://iridia.ulb.ac.be/~fmascia/maximum_clique/DIMACS-benchmark)

## Background Reading
- [Graph_theory](https://en.wikipedia.org/wiki/Graph_theory)
- [Graph Theory - Fundamentals](https://www.tutorialspoint.com/graph_theory/graph_theory_fundamentals.htm)
- [Dense Graph](https://en.wikipedia.org/wiki/Dense_graph)

