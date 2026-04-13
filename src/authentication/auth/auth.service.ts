import { Injectable, NotFoundException, UnauthorizedException } from '@nestjs/common';
import { UserService } from '../../modules/user/user.service';
import { BadRequestException, ConflictException } from '../../common/exception';
import * as argon from 'argon2';
import { v4 as uuid4 } from 'uuid';
import dayjs from 'dayjs';
import { UserLogin } from '../../interfaces';
import { RegisterDTO, ResetPasswordDTO } from './dto';
import { MailService } from '../mail/mail.service';
import { PrismaService } from '../../prisma/prisma.service';
import { TokensService } from '../tokens/tokens.service';
import { Prisma, Role } from '@prisma/client';

@Injectable()
export class AuthService {
    constructor(
        private readonly userService: UserService,
        private readonly mailService: MailService,
        private prisma: PrismaService,
        private tokensService: TokensService
    ) {}

    // Validate user for local strategy
    async validateUser(email: string, password: string): Promise<UserLogin> {
        const user = await this.userService.findUserByEmail(email);
        if (!user || !user.password || !user.isVerified)
            throw new BadRequestException('Tài khoản không tồn tại hoặc chưa được xác thực');
        const ok = await argon.verify(user.password, password);
        if (!ok) throw new BadRequestException('Sai mật khẩu');
        return {
            id: user.id,
            email: user.email,
            username: user.username,
            firstName: user.firstName,
            lastName: user.lastName,
            role: user.role
        };
    }

    // Register new user
    async register(dto: RegisterDTO ) : Promise<void>{
        // Kiểm tra email đã tồn tại chưa
        const existingUser = await this.userService.isEmailExist(dto.email);
        if (existingUser) throw new ConflictException('Email đã được đăng ký!');

        const token = uuid4();
        const code_expired = dayjs().add(30, 'minutes').toDate();
        const hashPassword = await argon.hash(dto.password);
        const username = await this.userService.generateUniqueUsernameFromEmail(dto.email);

        // Insert vào Database
        try {
            await this.prisma.user.create({
                data: {
                    email: dto.email,
                    password: hashPassword,
                    firstName: dto.firstName,
                    lastName: dto.lastName,
                    username,
                    role: Role.USER,
                    code: token,
                    expiresAt: code_expired
                },
            });
        } catch (error) {
            if (error instanceof Prisma.PrismaClientKnownRequestError && error.code === 'P2002') {
                const target = (error.meta?.target ?? []) as string[];
                if (target.includes('email')) {
                    throw new ConflictException('Email đã được đăng ký!');
                }
                if (target.includes('username')) {
                    throw new ConflictException('Tên người dùng đã tồn tại. Vui lòng thử lại.');
                }
            }
            throw error;
        }

        const name = `${dto.firstName} ${dto.lastName}`;
        try {
            await this.mailService.sendVerificationEmail(dto.email, name, token);
        } catch (err) {
            console.error('Failed to send verification email', err);
            // Xóa user đã tạo nếu gửi email thất bại để user có thể đăng ký lại
            try {
                await this.prisma.user.delete({
                    where: {
                        email: dto.email
                    }
                });
            } catch (cleanupError) {
                console.error('Failed to cleanup user after email send failure', cleanupError);
            }
            throw new BadRequestException('Đăng ký thất bại. Không thể gửi email xác thực, vui lòng thử lại sau.');
        }
    }

    // ĐĂNG NHẬP
    async login(user: UserLogin, device?: string, ip?: string) : Promise<{ access_token: string; refresh_token: string; user: UserLogin }> {
        const access_token = await this.tokensService.createAccessToken(user);
        const refresh_token = await this.tokensService.createRefreshToken(user);
        const hashed = await argon.hash(refresh_token);

        // Max 3 tokens logic
        const existingTokens = await this.prisma.refreshToken.findMany({
            where: { userId: user.id },
            orderBy: { createdAt: 'asc' }
        });

        if (existingTokens.length >= 3) {
            // Keep the latest 2 tokens to make room for 1 more
            const tokensToDelete = existingTokens.slice(0, existingTokens.length - 2);
            await this.prisma.refreshToken.deleteMany({
                where: {
                    id: { in: tokensToDelete.map(t => t.id) }
                }
            });
        }

        const expiresAt = dayjs().add(7, 'days').toDate(); // refresh token expiration

        await this.prisma.refreshToken.create({
            data: {
                token: hashed,
                device: device || null,
                ip: ip || null,
                expiresAt,
                userId: user.id
            }
        });

        return {
            access_token,
            refresh_token,
            user
        };
    }

    // LOGOUT
    async logout(email: string, refresh_token?: string) : Promise<void> {
        const user = await this.userService.findUserByEmail(email);
        if (!user) throw new NotFoundException('User not found');

        if (refresh_token) {
            // Xóa refresh token cụ thể nếu cung cấp (đăng xuất thiết bị hiện tại)
            const tokens = await this.prisma.refreshToken.findMany({
                where: { userId: user.id }
            });
            for (const token of tokens) {
                const isValid = await argon.verify(token.token, refresh_token);
                if (isValid) {
                    await this.prisma.refreshToken.delete({ where: { id: token.id } });
                    break;
                }
            }
        } else {
            // Xóa tất cả refresh token (đăng xuất mọi thiết bị)
            await this.prisma.refreshToken.deleteMany({
                where: { userId: user.id }
            });
        }
    }

    // REFRESH TOKEN
    async postrefresh_token(refresh_token: string) : Promise<{ access_token: string; user: UserLogin }> {
        const payload = await this.tokensService.verifyRefreshToken(refresh_token);
        if (!payload) {
            throw new UnauthorizedException('Invalid or expired refresh token');
        }

        const userRecord = await this.userService.findUserByEmail(payload.email);
        if (!userRecord) {
            throw new UnauthorizedException('User not found');
        }

        // Find refresh token in DB for this user
        const existingTokens = await this.prisma.refreshToken.findMany({
            where: { userId: userRecord.id },
            orderBy: { createdAt: 'desc' }
        });

        let validTokenId = null;
        for (const t of existingTokens) {
            const isValid = await argon.verify(t.token, refresh_token);
            if (isValid) {
                validTokenId = t.id;
                break;
            }
        }

        if (!validTokenId) {
            throw new UnauthorizedException('Invalid or expired refresh token');
        }

        const tokenRecord = await this.prisma.refreshToken.findUnique({
            where: { id: validTokenId },
            include: { user: true }
        });

        if (!tokenRecord || !tokenRecord.user) {
            throw new UnauthorizedException('User not found or refresh token revoked');
        }

        if (dayjs().isAfter(dayjs(tokenRecord.expiresAt))) {
            await this.prisma.refreshToken.delete({ where: { id: tokenRecord.id } });
            throw new UnauthorizedException('Refresh token has expired');
        }

        const user: UserLogin = {
            id: tokenRecord.user.id,
            email: tokenRecord.user.email,
            username: tokenRecord.user.username,
            firstName: tokenRecord.user.firstName,
            lastName: tokenRecord.user.lastName,
            role: tokenRecord.user.role
        };

        const access_token = await this.tokensService.createAccessToken(user);

        return {
            access_token,
            user
        };
    }

    // XÁC THỰC EMAIL
    async verifyByToken(token: string, email: string) : Promise<void> {
        if (!token) throw new BadRequestException('Token missing');

        const user = await this.userService.findUserByEmail(email);
        if (!user) throw new BadRequestException('Invalid or expired token');

        if (user.code !== token) {
            throw new BadRequestException('Invalid verification token');
        }

        if (user.expiresAt && dayjs().isAfter(dayjs(user.expiresAt))) {
            await this.prisma.user.delete({ where: { email } });
            throw new BadRequestException('Verification expired. Please register again.');
        }

        await this.prisma.user.update({
            where: { email },
            data: {
                isVerified: true,
                code: null,
                expiresAt: null
            }
        });
    }

    // RESET MẬT KHẨU
    async resetPassword(body: ResetPasswordDTO) : Promise<void> {
        if(body.newPassword !== body.confirmPassword) {
            throw new BadRequestException('Mật khẩu mới và xác nhận mật khẩu không khớp!');
        }
        const user = await this.userService.findUserByEmail(body.email);
        if (!user) throw new NotFoundException('User not found or invalid code');
        if (user.code !== body.code) {
            throw new BadRequestException('Invalid reset code');
        }
        const hashPassword = await argon.hash(body.newPassword);

        await this.prisma.user.update({
            where: { id: user.id },
            data: {
                password: hashPassword,
                code: null,
                expiresAt: null,
            }
        });
    }

    // GỬI LẠI EMAIL XÁC THỰC
    async resendVerificationEmail(email: string) : Promise<void> {
        const user = await this.userService.findUserByEmail(email);
        if (!user) throw new BadRequestException('Email không tồn tại');
        if (user.isVerified) throw new BadRequestException('Email đã được xác thực');

        const token = uuid4();
        const code_expired = dayjs().add(30, 'minutes').toDate();

        await this.prisma.user.update({
            where: { id: user.id },
            data: {
                code: token,
                expiresAt: code_expired
            }
        });

        const name = `${user.firstName} ${user.lastName}`;
        await this.mailService.sendVerificationEmail(user.email, name, token);
    }

    // GỬI EMAIL ĐẶT LẠI MẬT KHẨU
    async sendPasswordResetEmail(email: string) : Promise<void> {
        const user = await this.userService.findUserByEmail(email);
        if (!user) throw new BadRequestException('Email không tồn tại');

        const codeID = uuid4();
        const code_expired = dayjs().add(30, 'minutes').toDate();

        await this.prisma.user.update({
            where: { id: user.id },
            data: {
                code: codeID,
                expiresAt: code_expired
            }
        });

        const name = `${user.firstName} ${user.lastName}`;
        await this.mailService.sendPasswordResetEmail(user.email, name, codeID);
    }
}
