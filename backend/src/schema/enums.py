import enum


class AuthStatus(str, enum.Enum):
    AUTHENTICATED = "authenticated"
    NEEDS_REGISTRATION = "needs_registration"


class GenderType(str, enum.Enum):
    MALE = "male"
    FEMALE = "female"
    NON_BINARY = "non_binary"
    OTHER = "other"
    UNSPECIFIED = "unspecified"


class BowStyleType(str, enum.Enum):
    RECURVE = "recurve"
    COMPOUND = "compound"
    BAREBOW = "barebow"
    LONGBOW = "longbow"


class SlotLetterType(str, enum.Enum):
    A = "A"
    B = "B"
    C = "C"
    D = "D"


class JWTAlgorithm(str, enum.Enum):
    """Supported JWT signing algorithms.

    HS256, HS384, HS512: HMAC with SHA-256/384/512 (symmetric, secret-based)
    RS256, RS384, RS512: RSA with SHA-256/384/512 (asymmetric, key pair)
    ES256, ES384, ES512: ECDSA with SHA-256/384/512 (asymmetric, key pair)
    """

    HS256 = "HS256"
    HS384 = "HS384"
    HS512 = "HS512"
    RS256 = "RS256"
    RS384 = "RS384"
    RS512 = "RS512"
    ES256 = "ES256"
    ES384 = "ES384"
    ES512 = "ES512"
