from enum import StrEnum


class AuthStatus(StrEnum):
    AUTHENTICATED = "authenticated"
    NEEDS_REGISTRATION = "needs_registration"


class GenderType(StrEnum):
    MALE = "male"
    FEMALE = "female"
    NON_BINARY = "non_binary"
    OTHER = "other"
    UNSPECIFIED = "unspecified"


class BowStyleType(StrEnum):
    RECURVE = "recurve"
    COMPOUND = "compound"
    BAREBOW = "barebow"
    LONGBOW = "longbow"


class SlotLetterType(StrEnum):
    A = "A"
    B = "B"
    C = "C"
    D = "D"


class JWTAlgorithm(StrEnum):
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


class WSContentType(StrEnum):
    SHOT_CREATED = "shot.created"
    SHOT_DELETED = "shot.deleted"
    ARROW_CREATED = "arrow.created"
    ARROW_DELETED = "arrow.deleted"


class FaceType(StrEnum):
    WA_40_FULL = "wa_40cm_full"
    WA_60_FULL = "wa_60cm_full"
    WA_80_FULL = "wa_80cm_full"
    WA_122_FULL = "wa_122cm_full"
    WA_80_6RINGS = "wa_80cm_6rings"
    WA_122_6RINGS = "wa_122cm_6rings"
    WA_40_TRIPLE_VERTICAL = "wa_40cm_triple_vertical"
    WA_60_TRIPLE_TRIANGULAR = "wa_60cm_triple_triangular"
    NONE = "none"
