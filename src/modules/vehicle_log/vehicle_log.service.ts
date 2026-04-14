import { LogType, SlotStatus } from '@prisma/client';
import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../prisma/prisma.service';

@Injectable()
export class VehicleLogService {
	constructor(private readonly prisma: PrismaService) {}

	async recordByStatusTransition(slotId: number, oldStatus: SlotStatus, newStatus: SlotStatus): Promise<void> {
		let type: LogType | null = null;

		if (oldStatus === SlotStatus.AVAILABLE && newStatus === SlotStatus.OCCUPIED) {
			type = LogType.IN;
		}

		if (oldStatus === SlotStatus.OCCUPIED && newStatus === SlotStatus.AVAILABLE) {
			type = LogType.OUT;
		}

		if (!type) {
			return;
		}

		await this.prisma.vehicleLog.create({
			data: {
				slotId,
				type,
			},
		});
	}

	async getLogsBySlotId(slotId: number) {
		return await this.prisma.vehicleLog.findMany({
			where: { slotId },
			orderBy: { createdAt: 'desc' },
		});
	}
}
