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

*/
package sudoku

//TODO: add much more documentation and examples.
