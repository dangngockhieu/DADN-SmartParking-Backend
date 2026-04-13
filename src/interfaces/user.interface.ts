import { Role } from "@prisma/client"

export interface UserPaginate {
  id: number,
  firstName: string,
  lastName: string,
  username: string,
  email: string,
  role: Role
}

export interface UserFindEmail {
  id: number,
  firstName: string,
  lastName: string,
  email: string,
  password: string,
  username: string,
  role: Role
  isVerified: boolean,
  code: string | null,
  expiresAt: Date | null
}