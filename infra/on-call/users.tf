resource "pagerduty_user" "casey" {
  name  = "Casey"
  email = "casey@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.fred
  id = "PCMD3EC"
}

resource "pagerduty_user" "fred" {
  name  = "Fred"
  email = "fred@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.harsh
  id = "PDYFXKD"
}

resource "pagerduty_user" "harsh" {
  name  = "Harsh"
  email = "harsh@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.jon
  id = "PZI8VCV"
}

resource "pagerduty_user" "jon" {
  name  = "Jon"
  email = "jon@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.jordan
  id = "PB88HGZ"
}

resource "pagerduty_user" "jordan" {
  name  = "Jordan Acosta"
  email = "jordan@nuon.co"
  role  = "owner"
}

import {
  to = pagerduty_user.nat
  id = "PZQS32P"
}

resource "pagerduty_user" "nat" {
  name  = "Nat"
  email = "nat@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.rob
  id = "P4LR5YY"
}

resource "pagerduty_user" "rob" {
  name  = "Rob"
  email = "rob@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.sam
  id = "P49IMWW"
}

resource "pagerduty_user" "sam" {
  name  = "Sam"
  email = "sam@nuon.co"
  role  = "admin"
}

import {
  to = pagerduty_user.tim
  id = "P1L5Z29"
}


resource "pagerduty_user" "tim" {
  name  = "Tim"
  email = "tim@nuon.co"
  role  = "admin"
}
