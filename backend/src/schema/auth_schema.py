from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, ConfigDict, EmailStr, Field, field_validator

from schema.archer_schema import ArcherDemographics, ArcherNamesOptional, ArcherRead
from schema.enums import AuthStatus


class AuthCreate(BaseModel):
    archer_id: UUID = Field(..., description="Owner archer UUID for this session")
    session_token_hash: bytes = Field(..., description="Hash (NOT plaintext) of session token")
    created_at: datetime = Field(..., description="Creation timestamp (UTC)")
    expires_at: datetime = Field(..., description="UTC expiry timestamp of the session")
    ua: str | None = Field(default=None, description="User-Agent string (optional)")
    ip_inet: str | None = Field(default=None, description="Client IP (normalized)")

    model_config = ConfigDict(title="Auth Session Create", extra="forbid")


class AuthFilter(BaseModel):
    session_token_hash: bytes | None = Field(
        default=None, description="Hash (NOT plaintext) of session token"
    )
    revoked_at: datetime | None = Field(default=None, description="Revocation timestamp (UTC)")


class AuthSet(BaseModel):
    revoked_at: datetime | None = Field(default=None, description="Revocation timestamp (UTC)")


class AuthUpdate(BaseModel):
    where: AuthFilter = Field(..., description="Filter criteria to select Auth Sessions to update")
    data: AuthSet = Field(..., description="Fields to update on the selected Auth Session")

    model_config = ConfigDict(title="Auth Session Update", extra="forbid")

    @field_validator("data")
    @classmethod
    def _validate_data_not_empty(cls, v: AuthSet) -> AuthSet:
        if len(v.model_fields_set) == 0:
            raise ValueError("data must set at least one field")
        return v

    @field_validator("where")
    @classmethod
    def _validate_where_has_id(cls, v: AuthFilter) -> AuthFilter:
        if len(v.model_fields_set) == 0:
            raise ValueError("where must set at least one field")
        elif v.session_token_hash is None:
            raise ValueError("where.session_token_hash must be provided")
        return v


class AuthRead(AuthCreate):
    auth_id: UUID = Field(..., description="Authentication Session identifier (UUID)")
    created_at: datetime = Field(..., description="Creation timestamp (UTC)")
    revoked_at: datetime | None = Field(default=None, description="Revocation timestamp if revoked")

    def get_id(self) -> UUID:
        return self.auth_id


class GoogleOneTapRequest(BaseModel):
    credential: str = Field(min_length=10, description="Google ID token from One Tap callback")
    model_config = ConfigDict(title="Google One Tap Request", extra="forbid")


class AuthAuthenticated(BaseModel):
    status: AuthStatus = Field(default=AuthStatus.AUTHENTICATED)
    access_token: str = Field(..., description="Bearer access token for API requests")
    expires_at: datetime = Field(..., description="Access token expiry timestamp (UTC)")
    archer: ArcherRead = Field(..., description="Authenticated archer profile")
    model_config = ConfigDict(title="Auth Authenticated Response", extra="forbid")


class AuthNeedsRegistration(BaseModel):
    status: AuthStatus = Field(default=AuthStatus.NEEDS_REGISTRATION)
    google_email: EmailStr = Field(..., description="Google account email")
    google_subject: str = Field(..., description="Google account subject identifier (sub)")
    given_name: str | None = Field(default=None, description="Given name from Google (optional)")
    family_name: str | None = Field(default=None, description="Family name from Google (optional)")
    given_name_provided: bool = Field(
        ..., description="Whether Google provided a non-empty given_name"
    )
    family_name_provided: bool = Field(
        ..., description="Whether Google provided a non-empty family_name"
    )
    picture_url: str | None = Field(default=None, description="Google profile picture URL")
    model_config = ConfigDict(title="Auth Needs Registration Response", extra="forbid")


class LogoutResponse(BaseModel):
    success: bool = Field(..., description="Indicates if logout succeeded")
    model_config = ConfigDict(title="Logout Response", extra="forbid")


class AuthRegistrationRequest(ArcherDemographics, ArcherNamesOptional, BaseModel):
    credential: str = Field(min_length=10, description="Google ID token from One Tap callback")
    model_config = ConfigDict(title="Auth Registration Request", extra="forbid")
