import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';
import { SlotHistoryResponseDTO } from '../../interfaces';

@Injectable()
export class SlotHistoryService {
    constructor(private prisma: PrismaService) {}

    // Get slot history by slot ID
    async getSlotHistoryBySlotId(slotId: number): Promise<SlotHistoryResponseDTO[]> {
        const history = await this.prisma.slotHistory.findMany({
            where: { slotId },
            orderBy: { createdAt: 'desc' },
            select: {
                id: true,
                slotId: true,
                oldDevice: true,
                newDevice: true,
                oldPort: true,
                newPort: true,
                action: true,
                createdAt: true,
                user: {
                    select: {
                        id: true,
                        email: true,
                    },
                },
            },
        });
        return history;
    }
}
