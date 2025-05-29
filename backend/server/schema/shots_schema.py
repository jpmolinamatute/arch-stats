from pydantic import BaseModel, ConfigDict


class ShotsCreate(BaseModel):
    model_config = ConfigDict(extra="forbid")


class ShotsUpdate(BaseModel):
    model_config = ConfigDict(extra="forbid")
