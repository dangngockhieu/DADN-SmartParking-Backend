import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';
import { UserAccount, UserLogin } from '../../interfaces';

@Injectable()
export class TokensService {
    constructor(
        private jwt: JwtService,
        private config: ConfigService,
        ) {}

    async createAccessToken(user: UserLogin) {
        const payload = { sub: user.id, email: user.email, role: user.role };
        const accessToken = await this.jwt.signAsync(payload, {
            secret: this.config.get<string>('JWT_SECRET'),
            expiresIn: this.config.get<string>('JWT_EXPIRED') as any,
        });
        return accessToken;

    }

    async createRefreshToken(user: UserLogin ) {
        const payload = { sub: user.id, email: user.email};
        const refreshToken = await this.jwt.signAsync(payload, {
            secret: this.config.get<string>('JWT_REFRESH_SECRET'),
            expiresIn: this.config.get<string>('REFRESH_EXPIRED') as any,
        });
        return refreshToken;
    }

    async verifyRefreshToken(token: string) {
        try {
            return await this.jwt.verifyAsync(token, {
                secret: this.config.get<string>('JWT_REFRESH_SECRET'),
            });
        } catch (error) {
            return null;
        }
    }
}
