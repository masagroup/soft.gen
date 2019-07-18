package soft.generator.go;

import soft.generator.common.Generator;

/**
 * Entry point of the 'Generate' generation module.
 * 
 * @generated NOT
 */
public class Generate extends Generator {
    /**
     * The name of the module.
     */
    public static final String MODULE_FILE_NAME = "/soft/generator/go/generate";

    /**
     * The name of the templates that are to be generated.
     */
    public static final String[] TEMPLATE_NAMES = { "generateModel" };

    /**
     * This can be used to launch the generation from a standalone application.
     * 
     * @param args Arguments of the generation.
     */
    public static void main(String[] args) {
        generate(args, MODULE_FILE_NAME, TEMPLATE_NAMES);
    }

}
