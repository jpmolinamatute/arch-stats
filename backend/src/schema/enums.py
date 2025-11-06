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
