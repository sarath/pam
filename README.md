[![Stories in Ready](https://badge.waffle.io/sarath/pam.png?label=ready&title=Ready)](https://waffle.io/sarath/pam)
pam
===

Portable application manager (windows, for now)


End Game: remove the need for installers in windows. allow portable application developers a marketplace. 

Installation
------------

* download the release
* unzip to a folder that doesnot need admin access to write - eg: c:\me\pam *Important: this will be PAM_HOME, choose carefully*

First run
--------

Run the following in command line

    cd c:\me\pam
    pam

* the folder that pam is running in (c:\me\pam) is set as %PAM_HOME% folder
* %PAM_PATH% = %PAM_HOME% env var is created, appended to %PATH%
* .pamrc.json: registry is set to %PAM_HOME%\.reg (unzipped folder contains this), can be changed
* .pamrc.json: download cache is set to %PAM_HOME%\.cache (unzipped folder contains this), can be changed


Installing packages
--------------

Run the following in command line

    pam install <package>
    <or>
    pam install gh:<package>

Where <package> is

* a valid .reg/<package>.pam.json <or>
* TODO github.com/<package>.pam.json eg: <package> = sarath/pam-reg/go


removing packages
--------------
TODO

registry format
------------
TODO

Uninstallation
--------------
Removing pam is easy

* delete PAM_HOME folder
* delete %PAM_HOME% env variable
* delete %PAM_PATH% env variable
