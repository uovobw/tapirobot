tapirobot
=========

Rewrite of the irc-only part of gotapiri using hellabot as irc bot and removing the ajaxchat integration

Installation
============

Compile the executable with

    go build

then copy and edit the configuration file from the repository root (the values should be self-explanatory)

Upgrading
=========

Since hellabot supports seamless upgrades, tapiribot should too, launch the new executable and 
the upgrade should not drop the connection
