import { Body, Controller, Post, Headers, Req, Res, UseGuards, Request, Get, Query, Patch } from '@nestjs/common';
import { AuthService } from './auth.service';
import { RegisterDTO, LoginDTO, ResetPasswordDTO } from './dto';
import { Response } from 'express';
import { Public } from './decorators/customize';
import { UnauthorizedException } from '../../common/exception';
import { ConfigService } from '@nestjs/config';
import * as fs from 'fs';
import * as path from 'path';
import * as hbs from 'handlebars';

const resolveVerifiedTemplatePath = () => {
  const candidates = [
    path.join(process.cwd(), 'dist', 'authentication', 'mail', 'templates', 'verified.hbs'),
    path.join(process.cwd(), 'dist', 'templates', 'verified.hbs'),
    path.join(process.cwd(), 'src', 'authentication', 'mail', 'templates', 'verified.hbs'),
    path.join(__dirname, '..', 'mail', 'templates', 'verified.hbs'),
  ];

  const matched = candidates.find((templatePath) => fs.existsSync(templatePath));
  return matched ?? candidates[candidates.length - 1];
};

@Controller('auth')
export class AuthController {
  constructor(
    private readonly authService: AuthService,
    private readonly config: ConfigService
  ) {}

  // REGISTER
    @Post('register')
    @Public()
    async register(@Body() body: RegisterDTO) {
      await this.authService.register(body);
      return {
        message: 'Đăng ký thành công. Vui lòng kiểm tra email để xác thực tài khoản.',
        data:{}
      };
    }

  // LOGIN
    @Post('login')
    @Public()
    async login(@Body() body: LoginDTO, @Req() req: any, @Res({ passthrough: true }) res: Response, @Headers('user-agent') userAgent: string) {
      const user = await this.authService.validateUser(body.email, body.password);

      const ip = req.ip || req.connection?.remoteAddress;
      const data = await this.authService.login(user, userAgent, ip);

      const isProd = this.config.get<string>('NODE_ENV') === 'production';
      if (req.cookies?.refresh_token) {
        res.clearCookie('refresh_token', {
          httpOnly: true,
          secure: isProd,
          sameSite: isProd ? 'none' : 'lax',
          path: '/',
          // domain: isProd ? '.techzone.vn' : undefined
        });
      }

      res.cookie('refresh_token', data.refresh_token, {
        httpOnly: true,
        secure: isProd,
        sameSite: isProd ? 'strict' : 'lax',
        maxAge: 7 * 24 * 60 * 60 * 1000, // 7 ngày
        path: '/',
        // domain: isProd ? '.techzone.vn' : undefined
      });

      return {
        message: 'Đăng nhập thành công',
        data: {
          access_token: data.access_token,
          user: data.user
        }
      };
    }

  // LOGOUT
    @Post('logout')
    async logout(@Req() req: any, @Res({ passthrough: true }) res: Response) {
      const email = req.user?.email;
      if (!email) {
        throw new UnauthorizedException('No authenticated user found');
      }

      const refresh_token = req.cookies?.refresh_token;

      await this.authService.logout(email, refresh_token);

      const isProd = this.config.get<string>('NODE_ENV') === 'production';
      if (req.cookies?.refresh_token) {
        res.clearCookie('refresh_token', {
          httpOnly: true,
          secure: isProd,
          sameSite: isProd ? 'none' : 'lax',
          path: '/',
          // domain: isProd ? '.techzone.vn' : undefined
        });
      }

      return {
        message: 'Đăng xuất thành công',
        data: {}
      };
    }

  // REFRESH TOKEN
    @Post('refresh-token')
    @Public()
    async refreshToken(@Req() req: any, @Res({ passthrough: true }) res: Response) {
      const refresh_token = req.cookies?.refresh_token;
      if (!refresh_token) {
        throw new UnauthorizedException('Missing refresh token');
      }

      const data = await this.authService.postrefresh_token(refresh_token);

      const isProd = this.config.get<string>('NODE_ENV') === 'production';
      // Set lại refresh token vào cookie để đảm bảo cookie luôn fresh
      res.cookie('refresh_token', refresh_token, {
        httpOnly: true,
        secure: isProd,
        sameSite: isProd ? 'strict' : 'lax',
        maxAge: 7 * 24 * 60 * 60 * 1000, // 7 ngày
        path: '/'
      });

      return {
        message: 'Làm mới token thành công',
        data: {
          access_token: data.access_token,
          user: data.user,
        }
      };
    }

    // VERIFY EMAIL
    @Get('verify')
    @Public()
    async verify(
      @Query('token') token: string,
      @Query('email') email: string,
      @Res() res: Response,
    ) {
      await this.authService.verifyByToken(token, email);

      const templatePath = resolveVerifiedTemplatePath();
      const templateSource = fs.readFileSync(templatePath, 'utf-8');
      const compiledTemplate = hbs.compile(templateSource);

      const html = compiledTemplate({
        year: new Date().getFullYear(),
      });

      res.send(html);
    }

    // RESEND VERIFY EMAIL
    @Post('resend')
    @Public()
    async resend(@Body('email') email: string) {
      await this.authService.resendVerificationEmail(email);
      return {
        message: 'Resend email successful',
        data:{}
      };
    }

    // SEND RESET PASSWORD
    @Post('send-reset-password')
    @Public()
    async sendResetPassword(@Body('email') email: string) {
      await this.authService.sendPasswordResetEmail(email);
      return {
        message: 'Password reset email sent successfully',
        data:{}
      };
    }

    // RESET PASSWORD
    @Patch('reset-password')
    @Public()
    async resetPassword(@Body() body: ResetPasswordDTO) {
      await this.authService.resetPassword(body);
      return {
        message: 'Password reset successfully',
        data:{}
      };
    }
}
