import { useEffect } from "react";

export interface AlertRequest {
  message: string;
  /** Optional heading shown above the message. */
  title?: string;
  /** Visual tone of the dialog. Defaults to "success". */
  variant?: "success" | "error";
}

interface Props {
  request: AlertRequest;
  onClose: () => void;
}

/**
 * AlertDialog is a modal popup that surfaces a single outcome message (e.g. a
 * successful LinkedIn publish or a friendly error) with one dismiss button.
 * Closes on Escape, a backdrop click, or the OK button.
 */
export function AlertDialog({ request, onClose }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const variant = request.variant ?? "success";
  const title = request.title ?? (variant === "error" ? "Something went wrong" : "Success");

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className="modal card"
        role="alertdialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
      >
        <div className={`alert-head ${variant}`}>
          <span className="alert-icon" aria-hidden="true">
            {variant === "error" ? "✕" : "✓"}
          </span>
          <h2>{title}</h2>
        </div>
        <p className="modal-message">{request.message}</p>
        <div className="modal-actions">
          <button autoFocus onClick={onClose}>
            OK
          </button>
        </div>
      </div>
    </div>
  );
}
