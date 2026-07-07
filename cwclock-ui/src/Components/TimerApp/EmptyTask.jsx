import React from 'react';
import EmptyState from '../common/EmptyState';
import { useI18n } from '../../i18n/I18nContext';

const EmptyTask = () => {
  const { t } = useI18n();
  return <EmptyState title={t('timeTracker.emptyTitle')} body={t('timeTracker.emptyBody')} />;
}

export default EmptyTask
