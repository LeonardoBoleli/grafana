// ConfigCard.tsx
import { css } from '@emotion/css';
import React from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import { Button, Icon, Stack, useStyles2 } from '@grafana/ui';

import { IrmCardConfiguration } from './ConfigureIRM';
import { ProgressBar, StepsStatus } from './ProgressBar';

interface ConfigCardProps {
  config: IrmCardConfiguration;
  handleActionClick: (id: number, isDone: boolean | undefined) => void;
}

export function ConfigCard({ config, handleActionClick }: ConfigCardProps) {
  const styles = useStyles2(getStyles);
  return (
    <Stack direction={'column'} gap={1} justifyContent={'space-around'}>
      <div className={styles.cardContent}>
        <Stack direction={'column'} gap={1}>
          <Stack direction="row" alignItems="center" justifyContent="space-between" gap={1}>
            <Stack direction={'row'} gap={1} alignItems={'center'}>
              {config.title}
              {config.titleIcon && <Icon name={config.titleIcon} />}
              {config.isDone && <Icon name="check-circle" color="green" size="lg" />}
            </Stack>
            {config.stepsDone && config.totalStepsToDo && (
              <Stack direction="row" gap={0.5}>
                <StepsStatus stepsDone={config.stepsDone} totalStepsToDo={config.totalStepsToDo} />
                complete
              </Stack>
            )}
          </Stack>
          <Stack direction={'column'}>
            {config.description}
            {config.stepsDone && config.totalStepsToDo && (
              <ProgressBar stepsDone={config.stepsDone} totalStepsToDo={config.totalStepsToDo} />
            )}
          </Stack>
        </Stack>
        <Stack direction={'row'} gap={1} justifyContent={'flex-start'} alignItems={'flex-end'}>
          <Button variant="secondary" onClick={() => handleActionClick(config.id, config.isDone)}>
            {config.actionButtonTitle}
          </Button>
        </Stack>
      </div>
    </Stack>
  );
}

const getStyles = (theme: GrafanaTheme2) => ({
  cardTitle: css({
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
  }),
  cardContent: css({
    background: theme.colors.background.secondary,
    padding: theme.spacing(2),
    borderRadius: theme.shape.radius.default,
    height: '100%',
    gap: theme.spacing(1),
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
  }),
});
