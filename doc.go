/*

sudoku is a package for generating, solving, and rating the difficulty of sudoku puzzles.
Notably, it is able to solve a puzzle like a human would, and aspires to provide
high-quality difficulty ratings for puzzles based on a model trained on hundreds
of thousands of solves by real users.

The primary types are Grid and MutableGrid, which are interfaces. The Grid
interface contains only read-only methods, and MutableGrid contains all of
Grid's methods, plus mutating methods. Under the covers, this indirection
allows expensive searching operations that require hundreds of grid allocations
to be much more efficient.

To get a MutableGrid version of a Grid, use Grid.MutableCopy. MutableGrids can
be used as-is wherever a Grid is required.

To load a grid, see Load, LoadSDK, and LoadSDKFromFile.

To generate a new puzzle, see GenerateGrid.

To Solve a puzzle, see Grid.Solve.

To solve a puzzle like a human would, see Grid.HumanSolve.

*/
package sudoku

//TODO: add much more documentation and examples.
