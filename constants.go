package main

import "time"

const K = 2
const JOIN_MSG = "join"
const LIST_MSG = "list"
const PING_MSG = "ping"
const STORE_MSG = "store"
const FVALUE_MSG = "fValue"
const FVALUEFAIL_MSG = "fValueF"
const FLOOKUP_MSG = "fNode"
const ID_LENGTH = 20

// Server listen port
const RECEIVER_PORT = 1053
const TIMEOUT_DURATION = 2 * time.Second

const JOIN_WAIT_DURATION = 1 * time.Hour

var REPUBLISHED_DURATION = 10 * time.Second //1 * time.Hour
var EXPIRY_DURATION = 24 * time.Hour
