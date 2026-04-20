import { Body, Controller, Get, Post } from '@nestjs/common';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';
import { StatisticsService } from './statistics.service';
import { DateQueryDTO } from './dto/statistics.dto';
import { Roles } from '../../common/decorators/roles';
import { Role } from '@prisma/client';

@ApiTags('Statistics')
@Controller('statistics')
export class StatisticsController {
  constructor(private readonly statisticsService: StatisticsService) {}

  @Post('total-vehicles-in')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getTotalVehiclesIn(@Body() dto: DateQueryDTO) {
    const totalVehiclesIn = await this.statisticsService.getTotalVehiclesIn(dto.date);
    return { 'total vehicles in': totalVehiclesIn };
  }

  @Post('total-vehicles-out')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getTotalVehiclesOut(@Body() dto: DateQueryDTO) {
    const totalVehiclesOut = await this.statisticsService.getTotalVehiclesOut(dto.date);
    return { 'total vehicles out': totalVehiclesOut };
  }

  @Get('current-vehicles-in-parking')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getCurrentVehiclesInParking() {
    const currentVehiclesInParking = await this.statisticsService.getCurrentVehiclesInParking();
    return { 'current vehicles in parking': currentVehiclesInParking };
  }

  @Post('hourly-traffic')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getHourlyTraffic(@Body() dto: DateQueryDTO) {
    const hourlyTraffic = await this.statisticsService.getHourlyTraffic(dto.date);
    return { 'hourly traffic': hourlyTraffic };
  }

  @Post('peak-hour')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getPeakHour(@Body() dto: DateQueryDTO) {
    const { peakHour, totalVehiclesInPeakHour } = await this.statisticsService.getPeakHour(dto.date);
    return {
      'peak hour': peakHour,
      'total vehicles in peak hour': totalVehiclesInPeakHour,
    };
  }

  @Get('occupancy-rate')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getOccupancyRate() {
    const { occupancyRate, occupiedSlots, totalSlots } = await this.statisticsService.getOccupancyRate();
    return {
      'occupancy rate': occupancyRate,
      'occupied slots': occupiedSlots,
      'total slots': totalSlots,
    };
  }

  @Get('parking-status')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getParkingStatus() {
    const { status, statusCode } = await this.statisticsService.getParkingStatus();
    return {
      status,
      'status code': statusCode,
    };
  }
}
