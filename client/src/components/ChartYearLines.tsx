import { ReferenceLine } from "recharts";

export type ChartYearMarker = {
  value: string | number;
  year: number;
  separator: boolean;
};

type DatedPoint = { dt: string };
type TimedPoint = DatedPoint & { timestamp: number };

function pointYear(point: DatedPoint): number {
  return Number(point.dt.slice(0, 4));
}

export function categoricalYearMarkers(points: DatedPoint[]): ChartYearMarker[] {
  const markers = points.filter((point, index) => index === 0 || pointYear(point) !== pointYear(points[index - 1]));
  return markers.map((point, index) => ({ value: point.dt, year: pointYear(point), separator: index > 0 }));
}

export function timeYearMarkers(points: TimedPoint[]): ChartYearMarker[] {
  if (points.length === 0) return [];
  const firstYear = pointYear(points[0]);
  const lastYear = pointYear(points[points.length - 1]);
  const markers: ChartYearMarker[] = [{ value: points[0].timestamp, year: firstYear, separator: false }];
  for (let year = firstYear + 1; year <= lastYear; year += 1) {
    markers.push({ value: new Date(year, 0, 1).getTime(), year, separator: true });
  }
  return markers;
}

export function ChartYearLines({ markers, categorical = false }: { markers: ChartYearMarker[]; categorical?: boolean }) {
  return markers.map((marker) => (
    <ReferenceLine
      key={marker.year}
      label={{ value: String(marker.year), position: "insideTopLeft", fill: "#526963", fontSize: 12, fontWeight: 600 }}
      position={categorical ? "start" : undefined}
      stroke={marker.separator ? "#9bb1ab" : "transparent"}
      strokeDasharray="3 4"
      x={marker.value}
    />
  ));
}
