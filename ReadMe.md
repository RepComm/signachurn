# signachurn
if it ain't broke, don't rename it

## problem
naming things is hard, naming things twice is harder

at first glance, renaming after realising a change in scope seems like a good idea<br/>
but when you are someone's dependency, renaming causes many more issues.


## solution
signa-churn identifies projects with significant signature changes over commit/release history

dependents are encouraged to pick dependencies that do not have frequent renamings
dependencies are encouraged to pick better names, and adapt internals of existing code instead of reworking API entirely

## how
signachurn scans project trees, extracting all signatures and storing in a database

each tag (release) is compared to render information about changes to signatures over history

when a project renames a function, it will add a point to the project's local score

globally, dependent projects will multiply this score by their usage

the more API surface changes, the higher the score
and like golf, higher is worse

## under-the-hood
signachurn is a tool built using golang made of several components

- a database (pocketbase)
- a web frontend (built with preact)
- a scalable job system
- a job scan plugin system

the database stores calculated signature statistics for repositories

the web front end renders and queries the database for developers to see

the job system enables resumable parallel calculation of signature statistics

the job scan plugin system enables future kinds of language scanning

## goals
- enable developers to see global API changes
- encourage less churn by selecting stable API surfaces over instable ones
