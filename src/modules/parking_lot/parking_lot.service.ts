import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';
import { CreateParkingLotDTO, UpdateParkingLotDTO } from './dto';
import { ParkingLotAll, ParkingLotWithStats } from '../../interfaces';
import { NotFoundException } from '../../common/exception';

@Injectable()
export class ParkingLotService {
    constructor(private prisma: PrismaService) {}

    async getAllParkingLot() : Promise<ParkingLotAll[]> {
        const parkingLots = await this.prisma.parkingLot.findMany({
            select: {
                id: true,
                name: true,
                location: true,
            },
        });

        return parkingLots;

    }

    async getParkingLotById(id: number): Promise<ParkingLotWithStats> {
        const [parkingLot, statsRaw] = await Promise.all([
            this.prisma.parkingLot.findUnique({
            where: { id: id },
            select: {
                id: true,
                name: true,
                location: true,
                slots: {
                select: {
                    id: true,
                    name: true,
                    status: true,
                    deviceMac: true,
                    portNumber: true
                },
                },
            },
            }),

            this.prisma.parkingSlot.groupBy({
                by: ['status'],
                where: { lotId: id },
                _count: { status: true },
            }),
        ]);

        if (!parkingLot) {
            throw new NotFoundException('Parking lot not found');
        }

        // format stats
        const stats = {
            total: 0,
            available: 0,
            occupied: 0,
            maintain: 0,
        };

        statsRaw.forEach(item => {
            const count = item._count.status;
            stats.total += count;

            if (item.status === 'AVAILABLE') stats.available = count;
            if (item.status === 'OCCUPIED') stats.occupied = count;
            if (item.status === 'MAINTAIN') stats.maintain = count;
        });

        return {
            ...parkingLot,
            stats,
        };
    }

    async createParkingLot(dto: CreateParkingLotDTO): Promise<ParkingLotAll> {
        const createdLot = await this.prisma.parkingLot.create({
            data: {
                name: dto.name,
                location: dto.location,
            },
        });
        return createdLot;
    }

    async updateParkingLot(id: number, dto: UpdateParkingLotDTO): Promise<ParkingLotAll> {
        const data = Object.fromEntries(
            Object.entries(dto).filter(([_, v]) => v !== null && v !== undefined)
        );
        const updatedLot = await this.prisma.parkingLot.update({
            where: { id: id },
            data,
                select: {
                id: true,
                name: true,
                location: true,
            },
        });
        return updatedLot;
    }
}
