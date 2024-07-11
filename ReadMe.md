# depshit - API stability metrics
depshit analyzes libraries to rank them on API stability

## the problem
Inspired by watching [Linus Torvalds on why desktop Linux sucks - gentooman](https://www.youtube.com/watch?v=Pzl1B7nB9Kc)

When deciding what dependency we want to use, we use a metric: commit frequency.

Repo commit frequency is similar to dependability.

On it's own, commit frequency is really just commit churn. You can inflate trust by commiting lots of changes.

Frequent commits can incur frequent API changes.
A packager's nightmare.

Example of breaking changes:
```diff
+ //GetPerson by email changed to make more sense, see GetPersonByEmail
- function GetPerson ( email: string ): Person
+ function GetPerson ( id   : number ): Person
+ ...
+ function GetPersonByEmail( email: string ): Person
```
While this makes sense to the dev, it breaks any down-stream dependants, triggering upgrades and rebuilds.

## a solution
Give developers the metric of API stability, not just maintainer engagement

Bonus points: provide suggestions for dependency swapping / code changes required to migrate