import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';
import { LogType, SlotStatus } from '@prisma/client';

type HourlyTrafficItem = {
  time: string;
  'vehicles entering': number;
  'vehicles leaving': number;
};

@Injectable()
export class StatisticsService {
  constructor(private readonly prisma: PrismaService) {}

  private buildDayRange(date: string) {
    const start = new Date(`${date}T00:00:00`);
    const end = new Date(`${date}T23:59:59.999`);
    return { start, end };
  }

  async getTotalVehiclesIn(date: string): Promise<number> {
    const { start, end } = this.buildDayRange(date);
    return this.prisma.vehicleLog.count({
      where: {
        type: LogType.IN,
        createdAt: { gte: start, lte: end },
      },
    });
  }

  async getTotalVehiclesOut(date: string): Promise<number> {
    const { start, end } = this.buildDayRange(date);
    return this.prisma.vehicleLog.count({
      where: {
        type: LogType.OUT,
        createdAt: { gte: start, lte: end },
      },
    });
  }

  async getCurrentVehiclesInParking(): Promise<number> {
    return this.prisma.parkingSlot.count({
      where: { status: SlotStatus.OCCUPIED },
    });
  }

  async getHourlyTraffic(date: string): Promise<HourlyTrafficItem[]> {
    const { start, end } = this.buildDayRange(date);
    const logs = await this.prisma.vehicleLog.findMany({
      where: { createdAt: { gte: start, lte: end } },
      select: {
        createdAt: true,
        type: true,
      },
    });

    const traffic: HourlyTrafficItem[] = Array.from({ length: 24 }, (_, index) => ({
      time: `${index.toString().padStart(2, '0')}:00`,
      'vehicles entering': 0,
      'vehicles leaving': 0,
    }));

    logs.forEach((log) => {
      const hour = log.createdAt.getHours();
      const bucket = traffic[hour];
      if (log.type === LogType.IN) {
        bucket['vehicles entering'] += 1;
      } else {
        bucket['vehicles leaving'] += 1;
      }
    });

    return traffic;
  }

  async getPeakHour(date: string): Promise<{ peakHour: string; totalVehiclesInPeakHour: number }> {
    const traffic = await this.getHourlyTraffic(date);
    const peak = traffic.reduce(
      (best, current) => {
        const currentTotal = current['vehicles entering'] + current['vehicles leaving'];
        if (currentTotal > best.total) {
          return { hour: current.time, total: currentTotal };
        }
        return best;
      },
      { hour: '00:00', total: 0 },
    );

    return {
      peakHour: peak.hour,
      totalVehiclesInPeakHour: peak.total,
    };
  }

  async getOccupancyRate(): Promise<{ occupiedSlots: number; totalSlots: number; occupancyRate: string }> {
    const [occupiedSlots, totalSlots] = await this.prisma.$transaction([
      this.prisma.parkingSlot.count({ where: { status: SlotStatus.OCCUPIED } }),
      this.prisma.parkingSlot.count(),
    ]);
    const rate = totalSlots === 0 ? 0 : +(occupiedSlots / totalSlots) * 100;
    return {
      occupiedSlots,
      totalSlots,
      occupancyRate: `${rate.toFixed(2)}%`,
    };
  }

  async getParkingStatus(): Promise<{ status: string; statusCode: number }> {
    const { occupiedSlots, totalSlots } = await this.getOccupancyRate();
    const numericRate = totalSlots === 0 ? 0 : (occupiedSlots / totalSlots) * 100;

    if (numericRate >= 100) {
      return { status: 'Full', statusCode: 2 };
    }

    if (numericRate >= 80) {
      return { status: 'Busy', statusCode: 1 };
    }

    return { status: 'Available', statusCode: 0 };
  }
}
