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


class TargetFaceType(str, enum.Enum):
    CM_40_FULL = "40cm_full"
    CM_60_FULL = "60cm_full"
    CM_80_FULL = "80cm_full"
    CM_122_FULL = "122cm_full"
    CM_40_6RINGS = "40cm_6rings"
    CM_60_6RINGS = "60cm_6rings"
    CM_80_6RINGS = "80cm_6rings"
    CM_122_6RINGS = "122cm_6rings"
    CM_40_TRIPLE_VERTICAL = "40cm_triple_vertical"
    CM_60_TRIPLE_TRIANGULAR = "60cm_triple_triangular"
    NONE = "none"
