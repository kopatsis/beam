package config

import "time"

const MIN_ORDER_PRICE = 100

const PAGELEN int = 20
const ORDERLEN int = 6
const REVIEWLEN int = 16

const SHIPREQS int = 119
const IPREQS int = 59

const IPEREQS int = 119
const ESTREQS int = 199

const BATCH int = 45

const SHIPINTERVAL time.Duration = 60 * time.Second

const FAVES_LIMIT = 50
const SAVES_LIMIT = 15
const LAST_ORDERED_LIMIT = 50
const CUSTOM_LIST_LIMIT = 50
const MAX_CUSTOM_LISTS = 25

const GC_HANDLE string = "/giftcard"
const GC_IMG string = "https://cdn.com/gc_"
const GC_NAME string = "Gift Card"

const LOWER_INV = 150
const HIGHER_INV = 500
const LOWEST_INV = 75

const BASIS_PAGE = "https://domain.com"
const CONV_RATES = "https://api.fxratesapi.com/latest"

const VERIF_EXPIR_MINS = 60
const SIGNIN_EXPIR_MINS = 15
const TWOFA_EXPIR_MINS = 5
const MAX_TWOFA_ATTEMPTS = 3

const RESET_EXPIR_MINS = 60
const RESET_PASS_COOLDOWN = 3 // Days

const EMAIL_CHANGE_COOLDOWN = 7 // Days

const PASSWORD_MIN_CHARS = 12
const PASSWORD_MAX_CHARS = 256
const PASSWORD_MIN_SPECIALS = 1
const PASSWORD_MIN_LETTER = 1
const PASSWORD_MIN_NUMBERS = 1
const SPECIAL_CHAR_LIST = `!"#$%&'()*+,-./:;<=>?@[\]^_` + "`" + `{|}~`
const LETTER_LIST = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const NUMBER_LIST = "1234567890"

const LOGIN_ATTEMPTS_DEVICE = 4   // Per second
const LOGIN_ATTEMPTS_IP = 7       // Per second
const LOGIN_ATTEMPTS_ACCOUNT = 12 // Per second

const LOGIN_ATTEMPTS_MINUTE = 50
const LOGIN_ATTEMPTS_HOUR = 250

const LOCKOUT_MINUTES_DEVICE = 120
const LOCKOUT_MINUTES_IP = 120
const LOCKOUT_MINUTES_ACCOUNT = 30

const LOCKOUT_MINUTES_MINUTE = 120
const LOCKOUT_MINUTES_HOUR = 480

const CONFIRM_EMAIL_WAIT = 30     // seconds
const CONFIRM_EMAIL_MAX = 10      // attempts
const CONFIRM_EMAIL_COOLDOWN = 12 // hours

const SCHEDULED_INCOMPLETE_CUST = 60

const REVIEW_BUCKET_NAME = "reviews-storage-beam-a312"
