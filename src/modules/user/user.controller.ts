import { Body, Controller, Get, Param, ParseIntPipe, Patch, Post, Req } from '@nestjs/common';
import { Request } from "express";
import { UserService } from './user.service';
import { ChangePasswordDTO, ChangeRoleDTO, CreateUserDTO } from './dto';
import { Role } from '@prisma/client';
import { Roles } from '../../authentication/auth/decorators/roles';

@Controller('users')
export class UserController {
  constructor(private readonly userService: UserService) {}

  // Get User With Paginateion
  @Get()
  @Roles(Role.ADMIN, Role.MANAGER)
  async getUserWithPaginate( @Req() req: Request) {
    const { page, limit, search } = req.query;
    const pageNumber = Number(page) || 1;
    const limitNumber = Number(limit) || 10;

    const {users, total} = await this.userService.getUserWithPaginate(
        pageNumber,
        limitNumber,
        search ? String(search) : ''
    );

    return {
      message: 'Lấy danh sách người dùng thành công',
      data: {
        users: users,
        total: total
      }
    };
  }

  // Create User By Admin
  @Post()
  @Roles(Role.ADMIN)
  async createUserByAdmin(@Body() dto: CreateUserDTO) {
    await this.userService.createUserForAdmin(dto);
    return {
        message: 'Tạo người dùng thành công'
      };
  }

  // Change Password
  @Patch('change-password')
  async changePassword(@Req() req: Request, @Body() dto:ChangePasswordDTO) {
    const email = (req.user as any)?.email;
    await this.userService.changePassword(email, dto);
      return {
          message: 'Đổi mật khẩu thành công'
      };
  }

  // Change Role
  @Patch('change-role/:id')
  @Roles(Role.ADMIN)
  async changeRole(@Param('id', ParseIntPipe) id: number, @Body() dto: ChangeRoleDTO) {
    await this.userService.changeRole(id, dto);
      return {
          message: 'Đổi vai trò thành công'
      };
  }

}
