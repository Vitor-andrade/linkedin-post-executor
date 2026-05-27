import { useEffect } from "react";

export interface ConfirmRequest {
  message: string;
  confirmLabel?: string;
  /** When true, the confirm button uses the destructive (danger) style. */
  danger?: boolean;
  resolve: (ok: boolean) => void;
}

interface Props {
  request: ConfirmRequest;
  onClose: (ok: boolean) => void;
}

/**
 * ConfirmDialog is a modal replacement for window.confirm. It renders an
 * overlay with a message and Cancel/Confirm buttons, closes on Escape or a
 * backdrop click (treated as Cancel), and reports the choice via onClose.
 */
export function ConfirmDialog({ request, onClose }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="modal-overlay" onClick={() => onClose(false)}>
      <div
        className="modal card"
        role="dialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
      >
        <p className="modal-message">{request.message}</p>
        <div className="modal-actions">
          <button className="secondary" onClick={() => onClose(false)}>
            Cancel
          </button>
          <button
            className={request.danger ? "danger-btn" : ""}
            autoFocus
            onClick={() => onClose(true)}
          >
            {request.confirmLabel ?? "Confirm"}
          </button>
        </div>
      </div>
    </div>
  );
}
