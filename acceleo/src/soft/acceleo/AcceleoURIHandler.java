package soft.acceleo;

import java.util.Arrays;
import java.util.Stack;

import org.eclipse.acceleo.internal.parser.compiler.IAcceleoParserURIHandler;
import org.eclipse.emf.common.util.URI;

public class AcceleoURIHandler implements IAcceleoParserURIHandler {

	private static final String SEGMENT_EMPTY = "";
	private static final String SEGMENT_SELF = ".";
	private static final String SEGMENT_PARENT = "..";
	
	private String[] resolve( String[] segments ) {
		Stack<String> result = new Stack<String>();
		for ( String segment : segments ) {
			if ( SEGMENT_PARENT.equals(segment) ) {
				result.pop();
			}
			else if ( SEGMENT_SELF.equals(segment) ) {
				
			}
			else
				result.push(segment);
		}
		return (String[]) result.toArray(new String[result.size()]);
	}
	
	@Override
	public URI transform(URI uri) {
		URI newURI = uri;
		String[] segments = uri.segments();
		if ( Arrays.stream(segments).anyMatch(s -> SEGMENT_EMPTY.equals(s) || SEGMENT_SELF.equals(s) || SEGMENT_PARENT.equals(s) ))
			newURI = URI.createHierarchicalURI( uri.scheme(), uri.authority(), uri.device(), resolve(segments), uri.query(), uri.fragment());
		
		String uriString = newURI.toString();
		if (uriString.startsWith("jar:file:")) {
			int indexOf = uriString.indexOf(".jar!/");
			if (indexOf != -1) {
				String name = uriString;
				name = name.substring(0, indexOf);
				name = name.substring("jar:file:".length() + 1);
				if (name.endsWith("-SNAPSHOT")) {
					name = name.substring(0, name.length() - "-SNAPSHOT".length());
				}

				name = name.substring(0, name.lastIndexOf("-"));
				if (name.contains("/")) {
					name = name.substring(name.lastIndexOf("/"));
					name = name + "/";
				}
				name = "platform:/plugin" + name + uriString.substring(indexOf + ".jar!/".length());
				return URI.createURI(name);
			}
		}
		
		return newURI;
	}

}
