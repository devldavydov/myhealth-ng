import { useEffect } from "react";

export const NOTIFICATION_DURATION_MS = 4000;

type TimedNotificationProps = {
  duration?: number;
  message: string;
  onDismiss: () => void;
};

export function TimedNotification({
  duration = NOTIFICATION_DURATION_MS,
  message,
  onDismiss
}: TimedNotificationProps) {
  useEffect(() => {
    if (!message) return;

    const timeout = window.setTimeout(onDismiss, duration);
    return () => window.clearTimeout(timeout);
  }, [duration, message, onDismiss]);

  if (!message) return null;

  return (
    <div className="save-success timed-notification" role="status">
      <span>{message}</span>
      <svg
        aria-label="До автоматического скрытия"
        className="notification-timer"
        role="img"
        viewBox="0 0 20 20"
      >
        <circle className="notification-timer-track" cx="10" cy="10" r="8" />
        <circle
          className="notification-timer-progress"
          cx="10"
          cy="10"
          r="8"
          style={{ animationDuration: `${duration}ms` }}
        />
      </svg>
    </div>
  );
}
