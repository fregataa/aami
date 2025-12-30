-- AAMI Config Server - Add all_smi exporter type
-- Updated: 2025-12-31
-- Description: Add all_smi to exporter type CHECK constraint for multi-vendor AI accelerator monitoring
-- Reference: https://github.com/fregataa/aami/issues/2

-- Drop and recreate the CHECK constraint to include all_smi
ALTER TABLE exporters DROP CONSTRAINT IF EXISTS exporters_type_check;
ALTER TABLE exporters ADD CONSTRAINT exporters_type_check
    CHECK (type IN ('node_exporter', 'dcgm_exporter', 'all_smi', 'custom'));
