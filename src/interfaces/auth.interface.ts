import { Role } from "@prisma/client"

export interface UserAccount {
  id: number,
  email: string,
  role: Role,
}

export interface UserLogin{
  id: number,
  email: string,
  username: string,
  firstName: string,
  lastName: string,
  role: Role
}