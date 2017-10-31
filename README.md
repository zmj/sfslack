# sfslack

Hack week project 2016. A Slack extension for initiating ShareFile "send a file" and "request a file" workflows within a Slack conversation or channel. Later some refactoring was done towards production-readiness, but the project was deprioritized.

TODO features:
* Send files from ShareFile (folder picker web ui, authenticated web context)
* Other workflows (signature, approval, ?)
* Search
* Proxy notifications from other ShareFile services

TODO internal:
* Send in-channel replies through bot, ephemeral replies fall-back to bot DM
* Improve logging: plumb through context?
* Refactor service interfaces: server, workflow, sfauth, slackauth
* Persist authentication service stores?
