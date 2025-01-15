package config

import "time"

const PAGELEN int = 20
const ORDERLEN int = 6
const REVIEWLEN int = 16

const SHIPREQS int = 119
const IPREQS int = 59

const IPEREQS int = 119
const ESTREQS int = 199

const SHIPINTERVAL time.Duration = 60 * time.Second

const FAVES_LIMIT = 15
const SAVES_LIMIT = 50
const LAST_ORDERED_LIMIT = 50

const GC_HANDLE string = "/giftcard"
const GC_IMG string = "https://cdn.com/gc_"
const GC_NAME string = "Gift Card"
