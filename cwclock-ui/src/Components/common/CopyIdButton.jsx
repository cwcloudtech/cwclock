import React from "react";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import Tooltip from "./Tooltip";
import { useI18n } from "../../i18n/I18nContext";
import toastOptions from "../../Redux/toastOptions";

const CopyIdButton = ({ id, className }) => {
  const { t } = useI18n();

  if (!id) return null;

  const handleCopy = (e) => {
    e.preventDefault();
    e.stopPropagation();
    navigator.clipboard.writeText(id).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <Tooltip label={t("common.copyId")}>
      <button type="button" className={className} onClick={handleCopy}>
        <FaRegCopy style={{ fontSize: "16px" }} />
      </button>
    </Tooltip>
  );
};

export default CopyIdButton;
