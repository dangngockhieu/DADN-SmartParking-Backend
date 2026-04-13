import { ChangePasswordDTO, ChangeRoleDTO, CreateUserDTO } from './dto/global.dto';
import { Injectable } from '@nestjs/common';
import * as argon from 'argon2';
import { PrismaService } from '../../prisma/prisma.service';
import { UserFindEmail, UserPaginate } from '../../interfaces';
import { BadRequestException, ConflictException, NotFoundException } from '../../common/exception';

@Injectable()
export class UserService {
    constructor(private prisma: PrismaService) {}

    // Create UserName
    generateUsernameFromEmail(email: string): string {
        return email.split('@')[0];
    };

    // Check email exists
    async isEmailExist(email: string): Promise<boolean> {
        const user = await this.prisma.user.findUnique({
            where: { email },
        });
        return !!user;
    }

    // Find user by email
    async findUserByEmail(email: string): Promise<UserFindEmail> {
        const user = await this.prisma.user.findUnique({
            where: { email },
            select: {
                id: true,
                email: true,
                password: true,
                role: true,
                username: true,
                firstName: true,
                lastName: true,
                isVerified: true,
                code: true,
                expiresAt: true
            },
        });
        return user;
    }

    // Get User With Paginateion
    async getUserWithPaginate(page: number, limit: number, search?: string)
        : Promise<{users: UserPaginate[], total: number}>
    {
        const skip = (page - 1) * limit;

        // Kiểm tra nếu search thực sự có nội dung
        const searchFilter = search?.trim()
            ? {
                OR: [
                { firstName: { contains: search } },
                { lastName: { contains: search } },
                { email: { contains: search } },
                { username: { contains: search } },
                ],
            }
            : {};

        const [users, total] = await Promise.all([
            this.prisma.user.findMany({
            where: {
                isVerified: true,
                ...searchFilter,
            },
            skip,
            take: limit,
            orderBy: { id: 'asc' },
            select: {
                id: true,
                firstName: true,
                lastName: true,
                username: true,
                email: true,
                role: true
                },
            },
        ),

            this.prisma.user.count({
            where: {
                isVerified: true,
                ...searchFilter,
            },
            }),
        ]);

        return { users, total };
    }

    // Create User
    async createUserForAdmin (dto: CreateUserDTO): Promise<void> {
        const existingUser = await this.isEmailExist(dto.email);
        if (existingUser) throw new ConflictException('Email đã được đăng ký!');
        const hashPassword = await argon.hash(dto.password);
        await this.prisma.user.create({
            data: {
                email: dto.email,
                password: hashPassword,
                username: this.generateUsernameFromEmail(dto.email),
                firstName: dto.firstName,
                lastName: dto.lastName,
                role: dto.role
            },
        });
    };

    // Change Password
    async changePassword(userId: number, dto: ChangePasswordDTO): Promise<void> {
        if(dto.newPassword !== dto.confirmPassword){
            throw new BadRequestException('Mật khẩu mới và xác nhận mật khẩu không khớp!');
        }
        const user = await this.prisma.user.findUnique({
            where: { id: userId },
            select: { password: true },
        });

        if (!user) throw new NotFoundException('Người dùng không tồn tại!');

        const isOldPasswordValid = await argon.verify(user.password, dto.oldPassword);
        if (!isOldPasswordValid) throw new BadRequestException('Mật khẩu cũ không đúng!');

        const hashNewPassword = await argon.hash(dto.newPassword);
        await this.prisma.user.update({
            where: { id: userId },
            data: { password: hashNewPassword },
        });
    }

    // Change Role
    async changeRole(userId: number, dto: ChangeRoleDTO): Promise<void> {
        const user = await this.prisma.user.findUnique({
            where: { id: userId },
            select: {
                id: true,
                role: true
            },
        });

        if (!user) throw new NotFoundException('Người dùng không tồn tại!');

        await this.prisma.user.update({
            where: { id: userId },
            data: { role: dto.newRole },
        });
    }
}
