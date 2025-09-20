export interface CoordinateEvent {
  user_id: string;
  session_id: string;
  lat: number;
  lon: number;
  timestamp: string;
}

export interface LocationEvent {
  user_id: string;
  session_id: string;
  location: string;
  lat: number;
  lon: number;
  timestamp: string;
}

export interface UserMarker {
  userId: string;
  position: [number, number];
  lastLocation?: string;
  lastUpdate: string;
}
