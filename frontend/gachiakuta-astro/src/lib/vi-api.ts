import apiClient from './api-client';
import type { VitalInstrument } from '../types/vital-instruments';

export async function getAllVitalInstruments(): Promise<VitalInstrument[]> {
  return await apiClient.get<VitalInstrument[]>('/api/vital-instruments');
}

export async function getVitalInstrumentById(id: string): Promise<VitalInstrument> {
  return await apiClient.get<VitalInstrument>(`/api/vital-instruments/${id}`);
}

