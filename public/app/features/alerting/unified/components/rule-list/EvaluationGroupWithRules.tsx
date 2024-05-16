import { size } from 'lodash';
import React from 'react';
import { useToggle } from 'react-use';

import { createViewLink } from '../../utils/misc';
import { hashRulerRule } from '../../utils/rule-id';
import { isAlertingRule, isAlertingRulerRule, isGrafanaRulerRule, isRecordingRulerRule } from '../../utils/rules';

import { AlertRuleListItem, RecordingRuleListItem } from './AlertRuleListItem';
import EvaluationGroup from './EvaluationGroup';
import { EvaluationGroupWithRulesProps } from './RuleList.v2';

export const EvaluationGroupWithRules = ({ group, rulesSource }: EvaluationGroupWithRulesProps) => {
  const [open, toggleOpen] = useToggle(false);

  return (
    <EvaluationGroup name={group.name} interval={group.interval} isOpen={open} onToggle={toggleOpen}>
      {group.rules.map((rule) => {
        const { rulerRule, promRule, annotations } = rule;

        // keep in mind that we may not have a promRule for the ruler rule – this happens when the target
        // rule source is eventually consistent - it may know about the rule definition but not its state
        const isAlertingPromRule = isAlertingRule(promRule);

        if (isAlertingRulerRule(rulerRule)) {
          return (
            <AlertRuleListItem
              key={hashRulerRule(rulerRule)}
              state={isAlertingPromRule ? promRule?.state : undefined}
              health={promRule?.health}
              error={promRule?.lastError}
              name={rulerRule.alert}
              labels={rulerRule.labels}
              lastEvaluation={promRule?.lastEvaluation}
              evaluationDuration={promRule?.evaluationTime}
              evaluationInterval={group.interval}
              instancesCount={isAlertingPromRule ? size(promRule.alerts) : undefined}
              href={createViewLink(rulesSource, rule)}
              summary={annotations?.['summary']}
            />
          );
        }

        if (isRecordingRulerRule(rulerRule)) {
          return (
            <RecordingRuleListItem
              key={hashRulerRule(rulerRule)}
              name={rulerRule.record}
              health={promRule?.health}
              error={promRule?.lastError}
              lastEvaluation={promRule?.lastEvaluation}
              evaluationDuration={promRule?.evaluationTime}
              evaluationInterval={group.interval}
              labels={rulerRule.labels}
              href={createViewLink(rulesSource, rule)}
            />
          );
        }

        if (isGrafanaRulerRule(rulerRule)) {
          return (
            <AlertRuleListItem
              key={rulerRule.grafana_alert.uid}
              name={rulerRule.grafana_alert.title}
              state={isAlertingPromRule ? promRule?.state : undefined}
              labels={rulerRule.labels}
              health={promRule?.health}
              error={promRule?.lastError}
              isPaused={rulerRule.grafana_alert.is_paused}
              lastEvaluation={promRule?.lastEvaluation}
              evaluationDuration={promRule?.evaluationTime}
              evaluationInterval={group.interval}
              instancesCount={isAlertingPromRule ? size(promRule.alerts) : undefined}
              href={createViewLink(rulesSource, rule)}
              summary={rule.annotations?.['summary']}
              isProvisioned={Boolean(rulerRule.grafana_alert.provenance)}
            />
          );
        }

        return null;
      })}
    </EvaluationGroup>
  );
};
